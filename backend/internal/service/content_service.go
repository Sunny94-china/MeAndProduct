package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/entity"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/cache"
	"meandproduct/internal/pkg/slug"
	"meandproduct/internal/repository"
)

var (
	ErrFounderSlugConflict = errors.New("founder slug already in use")
	ErrFounderNotFound     = errors.New("founder profile not found")
	ErrProductNotFound     = errors.New("product not found")
	ErrTagNotFound         = errors.New("tag not found")
	ErrInvalidTarget       = errors.New("invalid target type or id")
	ErrInvalidReportStatus = errors.New("invalid report status transition")
	ErrReportNotFound      = errors.New("report not found")
	ErrTagSlugConflict     = errors.New("tag slug or name already exists")
)

// =========================================================
// FounderProfileService
// =========================================================

type FounderProfileService struct {
	userRepo *repository.UserRepo
	fpRepo   *repository.FounderProfileRepo
	linkRepo *repository.LinkRepo
	tagRepo  *repository.TagRepo
	rd       *redis.Client
}

func NewFounderProfileService(
	userRepo *repository.UserRepo,
	fpRepo *repository.FounderProfileRepo,
	linkRepo *repository.LinkRepo,
	tagRepo *repository.TagRepo,
	rd *redis.Client,
) *FounderProfileService {
	return &FounderProfileService{userRepo: userRepo, fpRepo: fpRepo, linkRepo: linkRepo, tagRepo: tagRepo, rd: rd}
}

func founderKey(id uint64) string         { return fmt.Sprintf("fp:id:%d", id) }
func founderBySlugKey(slug string) string { return fmt.Sprintf("fp:slug:%s", slug) }

func (s *FounderProfileService) GetByUserID(ctx context.Context, userID uint64) (*vo.FounderProfileVO, error) {
	if userID == 0 {
		return nil, ErrFounderNotFound
	}
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if fp == nil {
		return nil, ErrFounderNotFound
	}
	return s.buildFpVO(ctx, fp)
}

func (s *FounderProfileService) GetBySlug(ctx context.Context, slug string, includeHidden bool, viewer *uint64) (*vo.FounderProfileDetailVO, error) {
	if slug == "" {
		return nil, ErrFounderNotFound
	}
	var fp *entity.FounderProfile
	var err error
	if includeHidden {
		fp, err = cache.GetOrSet[entity.FounderProfile](ctx, s.rd, founderBySlugKey(slug)+":hidden", 60*time.Second,
			func() (*entity.FounderProfile, error) { return s.fpRepo.GetBySlug(ctx, slug, true) })
	} else {
		fp, err = cache.GetOrSet[entity.FounderProfile](ctx, s.rd, founderBySlugKey(slug), 5*60*time.Second,
			func() (*entity.FounderProfile, error) { return s.fpRepo.GetBySlug(ctx, slug, false) })
	}
	if err != nil {
		return nil, err
	}
	if fp == nil {
		return nil, ErrFounderNotFound
	}
	if !includeHidden && !fp.IsVisible {
		return nil, ErrFounderNotFound
	}
	if viewer == nil || *viewer != fp.UserID {
		_ = s.fpRepo.IncrViewCount(ctx, fp.ID, 1)
	}
	voFp, err := s.buildFpVO(ctx, fp)
	if err != nil {
		return nil, err
	}
	return &vo.FounderProfileDetailVO{FounderProfileVO: *voFp}, nil
}

func (s *FounderProfileService) Apply(ctx context.Context, userID uint64, req *dto.FounderApplyProfileReq) (*vo.FounderProfileVO, error) {
	existing, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	startSlug := req.Slug
	if !slug.IsValid(startSlug) {
		startSlug = slug.Generate(req.Slug, req.DisplayName)
	}
	selfID := uint64(0)
	if existing != nil {
		selfID = existing.ID
	}
	finalSlug, err := s.resolveSlugUnique(ctx, startSlug, selfID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		fp := &entity.FounderProfile{
			UserID:         userID,
			DisplayName:    req.DisplayName,
			Title:          emptyPtrIfEmpty(req.Title),
			Bio:            emptyPtrIfEmpty(req.Bio),
			Slug:           finalSlug,
			Location:       emptyPtrIfEmpty(req.Location),
			ContactEmail:   emptyPtrIfEmpty(req.ContactEmail),
			AvatarURL:      emptyPtrIfEmpty(req.AvatarURL),
			CoverURL:       emptyPtrIfEmpty(req.CoverURL),
			ProducthuntURL: emptyPtrIfEmpty(req.ProducthuntURL),
			WechatID:       emptyPtrIfEmpty(req.WechatID),
			BioLong:        emptyPtrIfEmpty(req.BioLong),
			WebsiteURL:     emptyPtrIfEmpty(req.WebsiteURL),
			IsVisible:      true,
		}
		out, err := s.txThenBuild(ctx, fp, req.CustomLinks, true)
		if err != nil {
			return nil, err
		}
		if err := s.replaceFounderTags(ctx, out.ID, req.TagIDs); err != nil {
			return nil, err
		}
		return s.buildFpVO(ctx, fp)
	}

	fields := map[string]any{
		"display_name":    req.DisplayName,
		"slug":            finalSlug,
		"title":           emptyPtrIfEmpty(req.Title),
		"bio":             emptyPtrIfEmpty(req.Bio),
		"location":        emptyPtrIfEmpty(req.Location),
		"contact_email":   emptyPtrIfEmpty(req.ContactEmail),
		"avatar_url":      emptyPtrIfEmpty(req.AvatarURL),
		"cover_url":       emptyPtrIfEmpty(req.CoverURL),
		"producthunt_url": emptyPtrIfEmpty(req.ProducthuntURL),
		"wechat_id":       emptyPtrIfEmpty(req.WechatID),
	}
	oldSlug := existing.Slug
	existing.Slug = finalSlug
	err = s.fpRepo.Tx(ctx, func(tx *gorm.DB) error {
		if err := s.fpRepo.UpdateByUserID(ctx, userID, fields); err != nil {
			return err
		}
		if req.CustomLinks != nil {
			if err := s.linkRepo.BatchUpsert(ctx, tx, entity.LinkSourceFounder, existing.ID, toRepoLinks(req.CustomLinks)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateFp(existing)
	if oldSlug != finalSlug {
		_ = cache.Invalidate(ctx, s.rd, founderBySlugKey(oldSlug), founderBySlugKey(oldSlug)+":hidden")
	}
	after, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.buildFpVO(ctx, after)
}

func (s *FounderProfileService) resolveSlugUnique(ctx context.Context, desiredSlug string, selfID uint64) (string, error) {
	candidate := desiredSlug
	for i := 0; i < 20; i++ {
		occ, err := s.userRepo.FounderSlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !occ {
			return candidate, nil
		}
		// Slug exists, but is it actually our own row?
		fp, err := s.fpRepo.GetBySlug(ctx, candidate, true)
		if err != nil {
			return "", err
		}
		if fp != nil && selfID > 0 && fp.ID == selfID {
			return candidate, nil
		}
		suffix := fmt.Sprintf("-%d", i+2)
		if len(desiredSlug)+len(suffix) > 50 {
			candidate = desiredSlug[:50-len(suffix)] + suffix
		} else {
			candidate = desiredSlug + suffix
		}
	}
	return "", ErrFounderSlugConflict
}

func (s *FounderProfileService) Update(ctx context.Context, userID uint64, role string, req *dto.FounderUpdateProfileReq) (*vo.FounderProfileVO, error) {
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if fp == nil {
		return nil, ErrFounderNotFound
	}
	fields := map[string]any{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Title != nil {
		fields["title"] = strPtrValue(*req.Title)
	}
	if req.Bio != nil {
		fields["bio"] = strPtrValue(*req.Bio)
	}
	if req.Location != nil {
		fields["location"] = strPtrValue(*req.Location)
	}
	if req.ContactEmail != nil {
		fields["contact_email"] = strPtrValue(*req.ContactEmail)
	}
	if req.AvatarURL != nil {
		fields["avatar_url"] = strPtrValue(*req.AvatarURL)
	}
	if req.CoverURL != nil {
		fields["cover_url"] = strPtrValue(*req.CoverURL)
	}
	if req.ProducthuntURL != nil {
		fields["producthunt_url"] = strPtrValue(*req.ProducthuntURL)
	}
	if req.WechatID != nil {
		fields["wechat_id"] = strPtrValue(*req.WechatID)
	}
	if req.BioLong != nil {
		fields["bio_long"] = strPtrValue(*req.BioLong)
	}
	if req.WebsiteURL != nil {
		fields["website_url"] = strPtrValue(*req.WebsiteURL)
	}
	if req.Slug != nil && *req.Slug != "" {
		finalSlug, err := s.resolveSlugUnique(ctx, *req.Slug, fp.ID)
		if err != nil {
			return nil, err
		}
		fields["slug"] = finalSlug
	}
	if req.IsVisible != nil && (role == string(entity.RoleAdmin) || *req.IsVisible) {
		fields["is_visible"] = *req.IsVisible
	}
	err = s.fpRepo.Tx(ctx, func(tx *gorm.DB) error {
		if len(fields) > 0 {
			if err := s.fpRepo.UpdateByUserID(ctx, userID, fields); err != nil {
				return err
			}
		}
		if req.CustomLinks != nil {
			if err := s.linkRepo.BatchUpsert(ctx, tx, entity.LinkSourceFounder, fp.ID, toRepoLinks(req.CustomLinks)); err != nil {
				return err
			}
		}
		if req.TagIDs != nil {
			if err := s.replaceFounderTagsTx(ctx, tx, fp.ID, *req.TagIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateFp(fp)
	after, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.buildFpVO(ctx, after)
}

func (s *FounderProfileService) List(ctx context.Context, filter dto.FounderListFilterReq) ([]*vo.FounderProfileItemVO, int64, error) {
	page := 1
	if filter.Page != nil {
		page = *filter.Page
	}
	pageSize := 20
	if filter.PageSize != nil {
		pageSize = *filter.PageSize
	}
	visible := true
	items, total, err := s.fpRepo.List(ctx, repository.FounderListFilter{
		TagIDs:    filter.TagIDs,
		IsVisible: &visible,
		SortBy:    filter.SortBy,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	result := make([]*vo.FounderProfileItemVO, 0, len(items))
	for _, it := range items {
		v, err := s.buildFpVO(ctx, it)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, &vo.FounderProfileItemVO{FounderProfileVO: *v})
	}
	return result, total, nil
}

func (s *FounderProfileService) txThenBuild(ctx context.Context, fp *entity.FounderProfile, links []dto.LinkRow, isApply bool) (*vo.FounderProfileVO, error) {
	err := s.fpRepo.Tx(ctx, func(tx *gorm.DB) error {
		if isApply {
			if err := s.fpRepo.Upsert(ctx, tx, fp); err != nil {
				return err
			}
		}
		if len(links) > 0 {
			if err := s.linkRepo.BatchUpsert(ctx, tx, entity.LinkSourceFounder, fp.ID, toRepoLinks(links)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateFp(fp)
	return s.buildFpVO(ctx, fp)
}

func (s *FounderProfileService) invalidateFp(fp *entity.FounderProfile) {
	if fp == nil {
		return
	}
	_ = cache.Invalidate(context.Background(), s.rd, founderKey(fp.ID), founderBySlugKey(fp.Slug),
		founderBySlugKey(fp.Slug)+":hidden")
}

func (s *FounderProfileService) buildFpVO(ctx context.Context, fp *entity.FounderProfile) (*vo.FounderProfileVO, error) {
	if fp == nil {
		return nil, ErrFounderNotFound
	}
	links, err := s.linkRepo.ListBySource(ctx, entity.LinkSourceFounder, fp.ID)
	if err != nil {
		return nil, err
	}
	tags, err := s.fpRepo.ListTags(ctx, fp.ID)
	if err != nil {
		return nil, err
	}
	return &vo.FounderProfileVO{
		ID:             fp.ID,
		UserID:         fp.UserID,
		DisplayName:    fp.DisplayName,
		Title:          ns1Value(fp.Title),
		Bio:            ns1Value(fp.Bio),
		BioLong:        fp.BioLong,
		Slug:           fp.Slug,
		Location:       ns1Value(fp.Location),
		AvatarURL:      ns1Value(fp.AvatarURL),
		CoverURL:       ns1Value(fp.CoverURL),
		ContactEmail:   ns1Value(fp.ContactEmail),
		WebsiteURL:     fp.WebsiteURL,
		ProducthuntURL: ns1Value(fp.ProducthuntURL),
		WechatID:       ns1Value(fp.WechatID),
		IsVisible:      fp.IsVisible,
		ViewCount:      fp.ViewCount,
		FavoriteCount:  fp.FavoriteCount,
		CustomLinks:    toVCLinks(links),
		Tags:           toTagVOList(tags),
		CreatedAt:      fp.CreatedAt,
		UpdatedAt:      fp.UpdatedAt,
	}, nil
}

func (s *FounderProfileService) replaceFounderTags(ctx context.Context, founderID uint64, tagIDs []uint64) error {
	return s.fpRepo.Tx(ctx, func(tx *gorm.DB) error {
		return s.replaceFounderTagsTx(ctx, tx, founderID, tagIDs)
	})
}

func (s *FounderProfileService) replaceFounderTagsTx(ctx context.Context, tx *gorm.DB, founderID uint64, tagIDs []uint64) error {
	tags, err := s.tagRepo.GetByIDs(ctx, tagIDs)
	if err != nil {
		return err
	}
	filtered := make([]uint64, 0, len(tags))
	for _, t := range tags {
		if t.IsActive {
			filtered = append(filtered, t.ID)
		}
	}
	added, removed, err := s.fpRepo.SetTagsReplace(ctx, tx, founderID, filtered)
	if err != nil {
		return err
	}
	for _, id := range added {
		_ = s.tagRepo.IncrFounderCount(ctx, id, 1)
	}
	for _, id := range removed {
		_ = s.tagRepo.IncrFounderCount(ctx, id, -1)
	}
	_ = cache.Invalidate(ctx, s.rd, "tags:active:list", "tags:grouped")
	return nil
}

// =========================================================
// ProductService
// =========================================================

type ProductService struct {
	productRepo *repository.ProductRepo
	fpRepo      *repository.FounderProfileRepo
	linkRepo    *repository.LinkRepo
	tagRepo     *repository.TagRepo
	rd          *redis.Client
}

func NewProductService(productRepo *repository.ProductRepo, fpRepo *repository.FounderProfileRepo, linkRepo *repository.LinkRepo, tagRepo *repository.TagRepo, rd *redis.Client) *ProductService {
	return &ProductService{productRepo: productRepo, fpRepo: fpRepo, linkRepo: linkRepo, tagRepo: tagRepo, rd: rd}
}

func productKey(id uint64) string { return fmt.Sprintf("prod:id:%d", id) }
func productBySlugKey(fSlug, pSlug string) string {
	return fmt.Sprintf("prod:slug:%s:%s", fSlug, pSlug)
}

func (s *ProductService) invalidateProduct(p *entity.Product) {
	if p == nil {
		return
	}
	_ = cache.Invalidate(context.Background(), s.rd, productKey(p.ID))
}

func (s *ProductService) GetByID(ctx context.Context, id uint64, includeHidden bool) (*vo.ProductVO, error) {
	if id == 0 {
		return nil, ErrProductNotFound
	}
	var p *entity.Product
	var err error
	if includeHidden {
		p, err = cache.GetOrSet[entity.Product](ctx, s.rd, productKey(id)+":hidden", 60*time.Second,
			func() (*entity.Product, error) { return s.productRepo.GetByID(ctx, id) })
	} else {
		p, err = cache.GetOrSet[entity.Product](ctx, s.rd, productKey(id), 5*60*time.Second,
			func() (*entity.Product, error) { return s.productRepo.GetByID(ctx, id) })
	}
	if err != nil {
		return nil, err
	}
	if p == nil || (!includeHidden && !p.IsVisible) {
		return nil, ErrProductNotFound
	}
	return s.buildProductVO(ctx, p)
}

func (s *ProductService) GetByFounderSlugs(ctx context.Context, founderSlug, productSlug string, includeHidden bool, viewer *uint64) (*vo.ProductDetailVO, error) {
	if founderSlug == "" || productSlug == "" {
		return nil, ErrProductNotFound
	}
	var p *entity.Product
	var err error
	if includeHidden {
		p, err = s.productRepo.GetByFounderSlugs(ctx, founderSlug, productSlug, true)
	} else {
		p, err = cache.GetOrSet[entity.Product](ctx, s.rd, productBySlugKey(founderSlug, productSlug), 5*60*time.Second,
			func() (*entity.Product, error) {
				return s.productRepo.GetByFounderSlugs(ctx, founderSlug, productSlug, false)
			})
	}
	if err != nil {
		return nil, err
	}
	if p == nil || (!includeHidden && !p.IsVisible) {
		return nil, ErrProductNotFound
	}
	if viewer == nil {
		_ = s.productRepo.IncrViewCount(ctx, p.ID, 1)
	} else {
		fp, _ := s.fpRepo.GetByID(ctx, p.FounderID)
		if fp == nil || fp.UserID != *viewer {
			_ = s.productRepo.IncrViewCount(ctx, p.ID, 1)
		}
	}
	voP, err := s.buildProductVO(ctx, p)
	if err != nil {
		return nil, err
	}
	_ = cache.Invalidate(ctx, s.rd, productKey(p.ID))
	return &vo.ProductDetailVO{ProductVO: *voP}, nil
}

func (s *ProductService) Create(ctx context.Context, userID uint64, req *dto.ProductCreateReq) (*vo.ProductVO, error) {
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if fp == nil {
		return nil, ErrFounderNotFound
	}
	startPslug := req.Slug
	if !validProductSlug2(startPslug) {
		startPslug = slug.Generate(req.Slug, req.Name)
	}
	if len(startPslug) > 200 {
		startPslug = startPslug[:200]
	}
	uniqSlug, err := slug.IsUnique(startPslug, func(try string) (bool, error) {
		existing, e := s.productRepo.GetByFounderSlugs(ctx, fp.Slug, try, true)
		return existing != nil, e
	})
	if err != nil {
		return nil, ErrProductNotFound
	}
	p := &entity.Product{
		FounderID:      fp.ID,
		Name:           req.Name,
		Slug:           uniqSlug,
		Tagline:        req.Tagline,
		Description:    req.Description,
		LogoURL:        req.LogoURL,
		CoverURL:       req.CoverURL,
		OfficialURL:    req.OfficialURL,
		GithubURL:      req.GithubURL,
		TwitterURL:     req.TwitterURL,
		ProducthuntURL: req.ProducthuntURL,
		PricingURL:     req.PricingURL,
		Status:         req.Status,
		IsVisible:      true,
	}
	if req.IsVisible != nil {
		p.IsVisible = *req.IsVisible
	}
	err = s.fpRepo.Tx(ctx, func(tx *gorm.DB) error {
		if err := s.productRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		if len(req.CustomLinks) > 0 {
			if err := s.linkRepo.BatchUpsert(ctx, tx, entity.LinkSourceProduct, p.ID, toRepoLinks(req.CustomLinks)); err != nil {
				return err
			}
		}
		if len(req.TagIDs) > 0 {
			tags, err := s.tagRepo.GetByIDs(ctx, req.TagIDs)
			if err != nil {
				return err
			}
			filtered := make([]uint64, 0, len(tags))
			for _, t := range tags {
				if t.IsActive {
					filtered = append(filtered, t.ID)
				}
			}
			added, _, err := s.productRepo.SetTagsReplace(ctx, tx, p.ID, filtered)
			if err != nil {
				return err
			}
			for _, id := range added {
				_ = s.tagRepo.IncrProductCount(ctx, id, 1)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateProduct(p)
	_ = cache.Invalidate(ctx, s.rd, "tags:active:list", "tags:grouped")
	return s.buildProductVO(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, userID uint64, role string, productID uint64, req *dto.ProductUpdateReq) (*vo.ProductVO, error) {
	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var check *uint64
	if role != string(entity.RoleAdmin) {
		if fp == nil || fp.ID != p.FounderID {
			return nil, ErrProductNotFound
		}
		check = &fp.ID
	}
	fields := map[string]any{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Slug != nil {
		startSlug := *req.Slug
		nameBase := p.Name
		if req.Name != nil {
			nameBase = *req.Name
		}
		if !validProductSlug2(startSlug) {
			startSlug = slug.Generate(*req.Slug, nameBase)
		}
		if len(startSlug) > 200 {
			startSlug = startSlug[:200]
		}
		newSlug, e := slug.IsUnique(startSlug, func(try string) (bool, error) {
			if try == p.Slug {
				return false, nil
			}
			existing, ee := s.productRepo.GetByFounderSlugs(ctx, "x", try, true)
			return existing != nil && existing.FounderID == p.FounderID, ee
		})
		if e != nil {
			return nil, ErrProductNotFound
		}
		fields["slug"] = newSlug
	}
	if req.Tagline != nil {
		fields["tagline"] = *req.Tagline
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.LogoURL != nil {
		fields["logo_url"] = strPtrValue(*req.LogoURL)
	}
	if req.CoverURL != nil {
		fields["cover_url"] = strPtrValue(*req.CoverURL)
	}
	if req.OfficialURL != nil {
		fields["official_url"] = strPtrValue(*req.OfficialURL)
	}
	if req.GithubURL != nil {
		fields["github_url"] = strPtrValue(*req.GithubURL)
	}
	if req.TwitterURL != nil {
		fields["twitter_url"] = strPtrValue(*req.TwitterURL)
	}
	if req.ProducthuntURL != nil {
		fields["producthunt_url"] = strPtrValue(*req.ProducthuntURL)
	}
	if req.PricingURL != nil {
		fields["pricing_url"] = strPtrValue(*req.PricingURL)
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.IsVisible != nil {
		fields["is_visible"] = *req.IsVisible
	}
	err = s.fpRepo.Tx(ctx, func(tx *gorm.DB) error {
		if len(fields) > 0 {
			if err := s.productRepo.Update(ctx, p.ID, fields, check); err != nil {
				return err
			}
		}
		if req.CustomLinks != nil {
			if err := s.linkRepo.BatchUpsert(ctx, tx, entity.LinkSourceProduct, p.ID, toRepoLinks(*req.CustomLinks)); err != nil {
				return err
			}
		}
		if req.TagIDs != nil {
			added, removed, err := s.productRepo.SetTagsReplace(ctx, tx, p.ID, *req.TagIDs)
			if err != nil {
				return err
			}
			for _, id := range added {
				_ = s.tagRepo.IncrProductCount(ctx, id, 1)
			}
			for _, id := range removed {
				_ = s.tagRepo.IncrProductCount(ctx, id, -1)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateProduct(p)
	_ = cache.Invalidate(ctx, s.rd, "tags:active:list", "tags:grouped")
	after, _ := s.productRepo.GetByID(ctx, productID)
	if after == nil {
		return nil, ErrProductNotFound
	}
	return s.buildProductVO(ctx, after)
}

func (s *ProductService) Delete(ctx context.Context, userID uint64, role string, productID uint64) error {
	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrProductNotFound
	}
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	var check *uint64
	if role != string(entity.RoleAdmin) {
		if fp == nil || fp.ID != p.FounderID {
			return ErrProductNotFound
		}
		check = &fp.ID
	}
	_, removed, err := s.productRepo.SetTagsReplace(ctx, nil, productID, nil)
	if err != nil {
		return err
	}
	for _, id := range removed {
		_ = s.tagRepo.IncrProductCount(ctx, id, -1)
	}
	if err := s.productRepo.SoftDelete(ctx, productID, check); err != nil {
		return err
	}
	_ = cache.Invalidate(ctx, s.rd, "tags:active:list", "tags:grouped")
	return nil
}

func (s *ProductService) ListMine(ctx context.Context, userID uint64, page, pageSize int) ([]*vo.ProductVO, int64, error) {
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if fp == nil {
		return []*vo.ProductVO{}, 0, nil
	}
	items, total, err := s.productRepo.ListByFounderID(ctx, fp.ID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*vo.ProductVO, 0, len(items))
	for _, it := range items {
		v, err := s.buildProductVO(ctx, it)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, nil
}

func (s *ProductService) GetMine(ctx context.Context, userID uint64, role string, productID uint64) (*vo.ProductVO, error) {
	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	if role != string(entity.RoleAdmin) {
		fp, err := s.fpRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if fp == nil || fp.ID != p.FounderID {
			return nil, ErrProductNotFound
		}
	}
	return s.buildProductVO(ctx, p)
}

func (s *ProductService) ListByFounder(ctx context.Context, founderSlug string, viewerUserID *uint64) ([]*vo.ProductVO, int64, error) {
	fp, err := s.fpRepo.GetBySlug(ctx, founderSlug, true)
	if err != nil || fp == nil {
		return []*vo.ProductVO{}, 0, ErrFounderNotFound
	}
	items, total, err := s.productRepo.ListByFounderID(ctx, fp.ID, 1, 500)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*vo.ProductVO, 0, len(items))
	for _, it := range items {
		if !it.IsVisible {
			if viewerUserID == nil {
				continue
			}
			fp2, _ := s.fpRepo.GetByID(ctx, it.FounderID)
			if fp2 == nil || fp2.UserID != *viewerUserID {
				continue
			}
		}
		v, err := s.buildProductVO(ctx, it)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, nil
}

func (s *ProductService) ListGlobal(ctx context.Context, filter dto.ProductListFilterReq) ([]*vo.ProductVO, int64, error) {
	page := 1
	if filter.Page != nil {
		page = *filter.Page
	}
	pageSize := 20
	if filter.PageSize != nil {
		pageSize = *filter.PageSize
	}
	visible := true
	items, total, err := s.productRepo.ListGlobal(ctx, repository.ProductListFilter{
		TagIDs:    filter.TagIDs,
		Stage:     filter.Stage,
		IsVisible: &visible,
		SortBy:    filter.SortBy,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*vo.ProductVO, 0, len(items))
	for _, it := range items {
		v, err := s.buildProductVO(ctx, it)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, nil
}

func (s *ProductService) Reorder(ctx context.Context, userID uint64, req *dto.ProductReorderReq) error {
	fp, err := s.fpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if fp == nil {
		return ErrFounderNotFound
	}
	return s.productRepo.Reorder(ctx, fp.ID, req.OrderedIDs)
}

func (s *ProductService) buildProductVO(ctx context.Context, p *entity.Product) (*vo.ProductVO, error) {
	links, err := s.linkRepo.ListBySource(ctx, entity.LinkSourceProduct, p.ID)
	if err != nil {
		return nil, err
	}
	tags, err := s.productRepo.ListTags(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return &vo.ProductVO{
		ID:             p.ID,
		FounderID:      p.FounderID,
		Name:           p.Name,
		Slug:           p.Slug,
		Tagline:        p.Tagline,
		Description:    p.Description,
		LogoURL:        ns1Value(p.LogoURL),
		CoverURL:       ns1Value(p.CoverURL),
		OfficialURL:    p.OfficialURL,
		GithubURL:      ns1Value(p.GithubURL),
		TwitterURL:     ns1Value(p.TwitterURL),
		ProducthuntURL: ns1Value(p.ProducthuntURL),
		PricingURL:     ns1Value(p.PricingURL),
		Status:         string(p.Status),
		IsVisible:      p.IsVisible,
		ViewCount:      p.ViewCount,
		FavoriteCount:  p.FavoriteCount,
		SortOrder:      p.SortOrder,
		CustomLinks:    toVCLinks(links),
		Tags:           toTagVOList(tags),
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}, nil
}

// =========================================================
// TagService
// =========================================================

type TagService struct {
	tagRepo *repository.TagRepo
	rd      *redis.Client
}

func NewTagService(tagRepo *repository.TagRepo, rd *redis.Client) *TagService {
	return &TagService{tagRepo: tagRepo, rd: rd}
}

func tagActiveKey() string  { return "tags:active:list" }
func tagGroupedKey() string { return "tags:grouped" }

func (s *TagService) ListActive(ctx context.Context) ([]*vo.TagVO, error) {
	list, err := cache.GetOrSet[[]entity.Tag](ctx, s.rd, tagActiveKey(), 5*60*time.Second,
		func() (*[]entity.Tag, error) {
			got, err := s.tagRepo.ListActive(ctx)
			if err != nil {
				return nil, err
			}
			out := make([]entity.Tag, 0, len(got))
			for _, t := range got {
				out = append(out, *t)
			}
			return &out, nil
		})
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []*vo.TagVO{}, nil
	}
	out := make([]*vo.TagVO, 0, len(*list))
	for i := range *list {
		out = append(out, toTagVO(&(*list)[i]))
	}
	return out, nil
}

func (s *TagService) GetByTypeGrouped(ctx context.Context) (map[string][]*vo.TagVO, error) {
	grouped, err := cache.GetOrSet[repository.GroupedTags](ctx, s.rd, tagGroupedKey(), 5*60*time.Second,
		func() (*repository.GroupedTags, error) {
			got, err := s.tagRepo.GetByTypeGrouped(ctx, false)
			if err != nil {
				return nil, err
			}
			return &got, nil
		})
	if err != nil {
		return nil, err
	}
	if grouped == nil {
		return map[string][]*vo.TagVO{}, nil
	}
	out := map[string][]*vo.TagVO{}
	for tp, list := range *grouped {
		sub := make([]*vo.TagVO, 0, len(list))
		for _, t := range list {
			sub = append(sub, toTagVO(t))
		}
		out[string(tp)] = sub
	}
	return out, nil
}

func (s *TagService) Create(ctx context.Context, req *dto.TagCreateReq) (*vo.TagVO, error) {
	t := &entity.Tag{
		Name:     req.Name,
		Slug:     req.Slug,
		Type:     entity.TagType(req.Type),
		IsActive: true,
	}
	if req.SortOrder != nil {
		t.SortOrder = *req.SortOrder
	}
	if err := s.tagRepo.Create(ctx, t); err != nil {
		return nil, ErrTagSlugConflict
	}
	s.invalidateTags()
	return toTagVO(t), nil
}

func (s *TagService) Update(ctx context.Context, id uint64, req *dto.TagUpdateReq) (*vo.TagVO, error) {
	fields := map[string]any{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Slug != nil {
		fields["slug"] = *req.Slug
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.SortOrder != nil {
		fields["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}
	if len(fields) > 0 {
		if err := s.tagRepo.Update(ctx, id, fields); err != nil {
			return nil, ErrTagSlugConflict
		}
	}
	s.invalidateTags()
	got, err := s.tagRepo.GetByIDs(ctx, []uint64{id})
	if err != nil || len(got) != 1 {
		return nil, ErrTagNotFound
	}
	return toTagVO(got[0]), nil
}

func (s *TagService) Deactivate(ctx context.Context, id uint64) error {
	if err := s.tagRepo.Deactivate(ctx, id); err != nil {
		return err
	}
	s.invalidateTags()
	return nil
}

func (s *TagService) invalidateTags() {
	_ = cache.Invalidate(context.Background(), s.rd, tagActiveKey(), tagGroupedKey())
}

func toTagVOList(tags []*entity.Tag) []vo.TagVO {
	out := make([]vo.TagVO, 0, len(tags))
	for _, t := range tags {
		if t == nil {
			continue
		}
		if v := toTagVO(t); v != nil {
			out = append(out, *v)
		}
	}
	return out
}

func toTagVO(t *entity.Tag) *vo.TagVO {
	return &vo.TagVO{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		Type:         string(t.Type),
		FounderCount: t.FounderCount,
		ProductCount: t.ProductCount,
		SortOrder:    t.SortOrder,
		IsActive:     t.IsActive,
	}
}

// =========================================================
// FavoriteService
// =========================================================

type FavoriteService struct {
	favRepo     *repository.FavoriteRepo
	fpRepo      *repository.FounderProfileRepo
	productRepo *repository.ProductRepo
}

func NewFavoriteService(favRepo *repository.FavoriteRepo, fpRepo *repository.FounderProfileRepo, productRepo *repository.ProductRepo) *FavoriteService {
	return &FavoriteService{favRepo: favRepo, fpRepo: fpRepo, productRepo: productRepo}
}

func (s *FavoriteService) resolveTarget(ctx context.Context, targetType entity.FavoriteTarget, targetID uint64) (bool, error) {
	switch targetType {
	case entity.FavoriteTargetFounder:
		fp, err := s.fpRepo.GetByIDV2(ctx, targetID)
		return err == nil && fp != nil, err
	case entity.FavoriteTargetProduct:
		p, err := s.productRepo.GetByID(ctx, targetID)
		return err == nil && p != nil, err
	}
	return false, ErrInvalidTarget
}

func (s *FavoriteService) Toggle(ctx context.Context, userID uint64, req *dto.FavoriteToggleReq) (vo.FavoriteToggleResult, error) {
	exists, err := s.resolveTarget(ctx, req.TargetType, req.TargetID)
	if err != nil {
		return vo.FavoriteToggleResult{}, err
	}
	if !exists {
		return vo.FavoriteToggleResult{}, ErrInvalidTarget
	}
	existed, err := s.favRepo.Upsert(ctx, userID, req.TargetType, req.TargetID)
	if err != nil {
		return vo.FavoriteToggleResult{}, err
	}
	var favoriteCount uint32
	if !existed {
		if err := s.favRepo.IncrTargetCount(ctx, nil, req.TargetType, req.TargetID, 1); err != nil {
			return vo.FavoriteToggleResult{}, err
		}
	} else {
		affected, err := s.favRepo.Delete(ctx, userID, req.TargetType, req.TargetID)
		if err != nil {
			return vo.FavoriteToggleResult{}, err
		}
		if affected > 0 {
			if err := s.favRepo.IncrTargetCount(ctx, nil, req.TargetType, req.TargetID, -1); err != nil {
				return vo.FavoriteToggleResult{}, err
			}
		}
	}
	switch req.TargetType {
	case entity.FavoriteTargetFounder:
		fp, _ := s.fpRepo.GetByIDV2(ctx, req.TargetID)
		if fp != nil {
			favoriteCount = fp.FavoriteCount
		}
	case entity.FavoriteTargetProduct:
		p, _ := s.productRepo.GetByID(ctx, req.TargetID)
		if p != nil {
			favoriteCount = p.FavoriteCount
		}
	}
	return vo.FavoriteToggleResult{
		TargetType:    string(req.TargetType),
		TargetID:      req.TargetID,
		Favorited:     !existed,
		FavoriteCount: favoriteCount,
	}, nil
}

func (s *FavoriteService) ListByUser(ctx context.Context, userID uint64, filter dto.FavoriteListReq) ([]*vo.FavoriteItemVO, int64, error) {
	page := 1
	if filter.Page != nil {
		page = *filter.Page
	}
	pageSize := 20
	if filter.PageSize != nil {
		pageSize = *filter.PageSize
	}
	items, total, err := s.favRepo.ListByUser(ctx, userID, repository.FavoriteListFilter{
		TargetType: filter.TargetType,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*vo.FavoriteItemVO, 0, len(items))
	for _, it := range items {
		item := &vo.FavoriteItemVO{
			ID:         it.ID,
			TargetType: string(it.TargetType),
			TargetID:   it.TargetID,
			CreatedAt:  it.CreatedAt,
		}
		switch it.TargetType {
		case entity.FavoriteTargetFounder:
			if fp, err := s.fpRepo.GetByIDV2(ctx, it.TargetID); err == nil && fp != nil {
				item.TargetSlug = fp.Slug
				item.TargetTitle = fp.DisplayName
				if fp.AvatarURL != nil && *fp.AvatarURL != "" {
					item.TargetCover = fp.AvatarURL
				}
				title := ""
				if fp.Title != nil {
					title = *fp.Title
				}
				bio := ""
				if fp.Bio != nil {
					bio = *fp.Bio
				}
				loc := ""
				if fp.Location != nil {
					loc = *fp.Location
				}
				avatar := ""
				if fp.AvatarURL != nil {
					avatar = *fp.AvatarURL
				}
				item.TargetSnapshot = map[string]any{
					"slug":         fp.Slug,
					"display_name": fp.DisplayName,
					"avatar_url":   avatar,
					"title":        title,
					"bio":          bio,
					"location":     loc,
				}
			}
		case entity.FavoriteTargetProduct:
			if p, err := s.productRepo.GetByID(ctx, it.TargetID); err == nil && p != nil {
				item.TargetSlug = p.Slug
				item.TargetTitle = p.Name
				if p.LogoURL != nil && *p.LogoURL != "" {
					item.TargetCover = p.LogoURL
				}
				fp, _ := s.fpRepo.GetByIDV2(ctx, p.FounderID)
				founderSlug := ""
				if fp != nil {
					founderSlug = fp.Slug
				}
				logo := ""
				if p.LogoURL != nil {
					logo = *p.LogoURL
				}
				item.TargetSnapshot = map[string]any{
					"slug":         p.Slug,
					"name":         p.Name,
					"tagline":      p.Tagline,
					"logo_url":     logo,
					"founder_slug": founderSlug,
				}
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *FavoriteService) Exists(ctx context.Context, userID uint64, req *dto.FavoriteToggleReq) (bool, error) {
	return s.favRepo.Exists(ctx, userID, req.TargetType, req.TargetID)
}

type SnapshotCountResult struct {
	FavoriteCount uint32
}

func (s *FavoriteService) SnapshotCount(ctx context.Context, targetType entity.FavoriteTarget, targetID uint64) (*SnapshotCountResult, error) {
	switch targetType {
	case entity.FavoriteTargetFounder:
		fp, err := s.fpRepo.GetByIDV2(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if fp == nil {
			return nil, nil
		}
		return &SnapshotCountResult{FavoriteCount: fp.FavoriteCount}, nil
	case entity.FavoriteTargetProduct:
		p, err := s.productRepo.GetByID(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, nil
		}
		return &SnapshotCountResult{FavoriteCount: p.FavoriteCount}, nil
	}
	return nil, nil
}

// =========================================================
// ReportService
// =========================================================

type ReportService struct {
	reportRepo  *repository.ReportRepo
	fpRepo      *repository.FounderProfileRepo
	productRepo *repository.ProductRepo
}

func NewReportService(r *repository.ReportRepo, fp *repository.FounderProfileRepo, p *repository.ProductRepo) *ReportService {
	return &ReportService{reportRepo: r, fpRepo: fp, productRepo: p}
}

func (s *ReportService) Create(ctx context.Context, reporterID *uint64, req *dto.ReportSubmitReq) (*vo.ReportVO, error) {
	ok, err := s.resolveTarget(ctx, req.TargetType, req.TargetID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidTarget
	}
	rep := &entity.Report{
		ReporterID:   reporterID,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		ReasonType:   entity.ReportReasonType(req.ReasonType),
		ReasonDetail: req.ReasonDetail,
		Status:       entity.ReportPending,
	}
	if req.ContactInfo != nil && *req.ContactInfo != "" {
		ci := *req.ContactInfo
		rep.ContactInfo = &ci
	}
	if err := s.reportRepo.Create(ctx, nil, rep); err != nil {
		return nil, err
	}
	return toReportVO(rep), nil
}

func (s *ReportService) List(ctx context.Context, filter dto.ReportListFilterReq) ([]*vo.ReportVO, int64, error) {
	page := 1
	if filter.Page != nil {
		page = *filter.Page
	}
	pageSize := 20
	if filter.PageSize != nil {
		pageSize = *filter.PageSize
	}
	items, total, err := s.reportRepo.List(ctx, repository.ReportListFilter{
		Status:     filter.Status,
		ReporterID: filter.ReporterID,
		TargetType: filter.TargetType,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*vo.ReportVO, 0, len(items))
	for _, it := range items {
		out = append(out, toReportVO(it))
	}
	return out, total, nil
}

func (s *ReportService) Handle(ctx context.Context, handlerID uint64, req *dto.ReportHandleReq) (*vo.ReportVO, error) {
	rp, err := s.reportRepo.GetByID(ctx, req.ReportID)
	if err != nil {
		return nil, err
	}
	if rp == nil {
		return nil, ErrReportNotFound
	}
	if rp.Status != entity.ReportPending {
		return nil, ErrInvalidReportStatus
	}
	_, _, err = s.reportRepo.Handle(ctx, nil, req.ReportID, repository.HandleReportPayload{
		HandlerID:   handlerID,
		NewStatus:   req.Status,
		HandlerNote: req.HandlerNote,
		HideTarget:  req.HideTarget,
	})
	if err != nil {
		return nil, err
	}
	after, _ := s.reportRepo.GetByID(ctx, req.ReportID)
	if after == nil {
		return nil, ErrReportNotFound
	}
	return toReportVO(after), nil
}

func (s *ReportService) resolveTarget(ctx context.Context, tt entity.ReportTarget, id uint64) (bool, error) {
	switch tt {
	case entity.ReportTargetFounder:
		fp, err := s.fpRepo.GetByIDV2(ctx, id)
		return err == nil && fp != nil, err
	case entity.ReportTargetProduct:
		p, err := s.productRepo.GetByID(ctx, id)
		return err == nil && p != nil, err
	}
	return false, ErrInvalidTarget
}

func toReportVO(r *entity.Report) *vo.ReportVO {
	v := &vo.ReportVO{
		ID:           r.ID,
		TargetType:   string(r.TargetType),
		TargetID:     r.TargetID,
		ReasonType:   string(r.ReasonType),
		ReasonDetail: r.ReasonDetail,
		Status:       string(r.Status),
		HideTarget:   r.HideTarget,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
	if r.ReporterID != nil && *r.ReporterID > 0 {
		v.ReporterID = r.ReporterID
	}
	if r.ContactInfo != nil {
		v.ContactInfo = *r.ContactInfo
	}
	if r.HandlerID != nil && *r.HandlerID > 0 {
		v.HandlerID = r.HandlerID
	}
	if r.HandlerNote != nil {
		v.HandlerNote = r.HandlerNote
	}
	if r.HandledAt != nil {
		v.HandledAt = r.HandledAt
	}
	return v
}

// =========================================================
// LinkService (track-click convenience endpoint)
// =========================================================

type LinkService struct {
	linkRepo    *repository.LinkRepo
	fpRepo      *repository.FounderProfileRepo
	productRepo *repository.ProductRepo
}

func NewLinkService(l *repository.LinkRepo, fp *repository.FounderProfileRepo, p *repository.ProductRepo) *LinkService {
	return &LinkService{linkRepo: l, fpRepo: fp, productRepo: p}
}

func (s *LinkService) ListBySource(ctx context.Context, sourceType entity.LinkSource, sourceID uint64) ([]*entity.Link, error) {
	return s.linkRepo.ListBySource(ctx, sourceType, sourceID)
}

func (s *LinkService) TrackClick(ctx context.Context, id uint64) (*entity.Link, error) {
	if id == 0 {
		return nil, ErrInvalidTarget
	}
	if _, err := s.linkRepo.IncrClick(ctx, id, 1); err != nil {
		return nil, err
	}
	return s.linkRepo.GetByID(ctx, id)
}

// =========================================================
// helpers
// =========================================================

func toRepoLinks(xs []dto.LinkRow) []repository.LinkRow {
	out := make([]repository.LinkRow, 0, len(xs))
	for _, x := range xs {
		out = append(out, repository.LinkRow{LinkType: x.LinkType, URL: x.URL})
	}
	return out
}

func emptyPtrIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strPtr(s string) *string {
	return &s
}

func nullStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullStrPtr2(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullStrP(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nsPValue(ns **string) *string {
	if ns == nil || *ns == nil {
		return nil
	}
	v := **ns
	if v == "" {
		return nil
	}
	return &v
}

func ns1Value(ns *string) *string {
	if ns == nil || *ns == "" {
		return nil
	}
	return ns
}

func ns2Value(ns **string) *string {
	if ns == nil || *ns == nil {
		return nil
	}
	v := **ns
	if v == "" {
		return nil
	}
	return &v
}

func str(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strPtrValue(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrStr(p **string) *string {
	if p == nil {
		return nil
	}
	return *p
}

func toVCLinks(links []*entity.Link) []vo.CustomLinkVO {
	out := make([]vo.CustomLinkVO, 0, len(links))
	for _, l := range links {
		out = append(out, vo.CustomLinkVO{
			ID:         l.ID,
			Label:      l.LinkType,
			LinkType:   l.LinkType,
			URL:        l.URL,
			ClickCount: l.ClickCount,
		})
	}
	return out
}

func validProductSlug2(s string) bool {
	return slug.IsValid(s) || (len(s) >= 3 && len(s) <= 200 && isAsciiSlug(s))
}

func validTagSlug(s string) bool {
	return slug.IsValid(s) || (len(s) >= 1 && len(s) <= 50 && isAsciiSlug(s))
}

func isAsciiSlug(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}
