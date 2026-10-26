package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"meandproduct/internal/config"
	"meandproduct/internal/handler"
	"meandproduct/internal/middleware"
	"meandproduct/internal/model/entity"
	"meandproduct/internal/pkg/cache"
	"meandproduct/internal/pkg/logger"
	migratepkg "meandproduct/internal/pkg/migrate"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/repository"
	"meandproduct/internal/service"
)

func main() {
	var (
		cfgPath       = flag.String("config", "", "optional path to config.yaml")
		doMigrateUp   = flag.Bool("migrate-up", false, "run all migrations then exit")
		doMigrateDn   = flag.Int("migrate-down", 0, "roll back N steps then exit")
		doEnsureAdmin = flag.Bool("ensure-admin", false, "upsert configured admin user then exit")
	)
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		panic("load config failed: " + err.Error())
	}

	lg := logger.New(cfg.AppEnv == "production")
	slog.SetDefault(lg)

	slog.Info("starting meandproduct backend", "env", cfg.AppEnv, "port", cfg.Server.Port)

	if *doMigrateUp || *doMigrateDn > 0 {
		if err := runMigrations(cfg, *doMigrateUp, *doMigrateDn); err != nil {
			slog.Error("migration failed", "err", err)
			os.Exit(1)
		}
		if *doMigrateUp {
			slog.Info("migrate-up finished")
		} else {
			slog.Info("migrate-down finished", "steps", *doMigrateDn)
		}
		return
	}

	db, err := initMySQL(cfg)
	if err != nil {
		slog.Error("mysql init failed", "err", err)
		os.Exit(1)
	}

	if *doEnsureAdmin {
		if cfg.MySQL.MigrateAuto {
			if err := runMigrations(cfg, true, 0); err != nil {
				slog.Error("auto-migrate failed", "err", err)
				os.Exit(1)
			}
		}
		if err := service.EnsureConfiguredAdmin(context.Background(), cfg, db); err != nil {
			slog.Error("ensure admin failed", "err", err)
			os.Exit(1)
		}
		slog.Info("configured admin ready", "username", cfg.Admin.Username)
		return
	}

	rd := initRedis(cfg)
	if err := rd.Ping(context.Background()).Err(); err != nil {
		slog.Error("redis ping failed", "err", err)
		os.Exit(1)
	}

	if cfg.MySQL.MigrateAuto {
		if err := runMigrations(cfg, true, 0); err != nil {
			slog.Error("auto-migrate failed", "err", err)
			os.Exit(1)
		}
	}

	if err := service.EnsureConfiguredAdmin(context.Background(), cfg, db); err != nil {
		slog.Error("ensure admin failed", "err", err)
		os.Exit(1)
	}

	router := buildRouter(cfg, db, rd)
	_ = db

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		slog.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown err", "err", err)
	}
	slog.Info("server exited")
}

func initMySQL(cfg *config.Config) (*gorm.DB, error) {
	gormLevel := gormlogger.Warn
	if cfg.AppEnv != "production" {
		gormLevel = gormlogger.Info
	}
	gl := gormlogger.New(
		slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  cfg.AppEnv != "production",
		},
	)

	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN()), &gorm.Config{Logger: gl})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
}

func buildRouter(cfg *config.Config, db *gorm.DB, rd *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.RateLimitGeneral(cfg, rd))

	// --- Wire deps ---
	userRepo := repository.NewUserRepo(db)
	fpRepo := repository.NewFounderProfileRepo(db)
	productRepo := repository.NewProductRepo(db)
	tagRepo := repository.NewTagRepo(db)
	favRepo := repository.NewFavoriteRepo(db)
	reportRepo := repository.NewReportRepo(db)
	linkRepo := repository.NewLinkRepo(db)

	authSvc := service.NewAuthService(cfg, db, rd, userRepo)
	authHdl := handler.NewAuthHandler(cfg, authSvc, rd)
	inviteSvc := service.NewInviteService(userRepo)
	inviteHdl := handler.NewInviteHandler(inviteSvc)

	fpSvc := service.NewFounderProfileService(userRepo, fpRepo, linkRepo, tagRepo, rd)
	productSvc := service.NewProductService(productRepo, fpRepo, linkRepo, tagRepo, rd)
	tagSvc := service.NewTagService(tagRepo, rd)
	favSvc := service.NewFavoriteService(favRepo, fpRepo, productRepo)
	reportSvc := service.NewReportService(reportRepo, fpRepo, productRepo)
	linkSvc := service.NewLinkService(linkRepo, fpRepo, productRepo)

	founderHdl := handler.NewFounderHandler(fpSvc)
	productHdl := handler.NewProductHandler(productSvc)
	tagHdl := handler.NewTagHandler(tagSvc)
	favHdl := handler.NewFavoriteHandler(favSvc)
	reportHdl := handler.NewReportHandler(reportSvc)
	linkHdl := handler.NewLinkHandler(linkSvc)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "ok",
				"data":    gin.H{"status": "UP", "env": cfg.AppEnv},
			})
		})

		// --- Auth (public) ---
		auth := v1.Group("/auth")
		{
			auth.POST("/register/guest", func(c *gin.Context) { authHdl.RegisterGuest(c) })
			auth.POST("/register/self", func(c *gin.Context) { authHdl.RegisterSelf(c) })
			auth.POST("/register/invite", func(c *gin.Context) { authHdl.RegisterInvite(c) })
			auth.POST("/login", middleware.RateLimitLogin(cfg, rd), func(c *gin.Context) { authHdl.Login(c) })
			auth.POST("/refresh", func(c *gin.Context) { authHdl.Refresh(c) })
			auth.POST("/invite/validate", func(c *gin.Context) { authHdl.ValidateInvite(c) })

			// --- Auth (protected) ---
			protected := auth.Group("")
			protected.Use(middleware.JWTAuth(cfg, rd))
			{
				protected.POST("/logout", func(c *gin.Context) { authHdl.Logout(c) })
				protected.GET("/me", func(c *gin.Context) { authHdl.Me(c) })
			}
		}

		// --- Public content ---
		public := v1.Group("/public")
		{
			// Founder
			public.GET("/founders", func(c *gin.Context) { founderHdl.PublicList(c) })
			public.GET("/founders/:slug", func(c *gin.Context) { founderHdl.PublicGetBySlug(c) })
			public.GET("/founders/:slug/products", func(c *gin.Context) { productHdl.PublicListByFounder(c) })
			public.GET("/founders/:slug/products/:p_slug", func(c *gin.Context) { productHdl.PublicGetByFounderSlugs(c) })

			// Products
			public.GET("/products", func(c *gin.Context) { productHdl.PublicListGlobal(c) })

			// Tags
			public.GET("/tags", func(c *gin.Context) { tagHdl.PublicListActive(c) })
			public.GET("/tags/grouped", func(c *gin.Context) { tagHdl.PublicGetGrouped(c) })

			// Link click tracking (no auth)
			public.POST("/links/:id/click", func(c *gin.Context) { linkHdl.TrackClick(c) })

			// Report submit (allow anonymous + report-level 5/IP/1h 限流)
			public.POST("/reports", middleware.RateLimitReport(cfg, rd), func(c *gin.Context) { reportHdl.Submit(c) })
		}

		// 保留：P1 期间测试 RBAC + JWTAuth 的占位
		testProtected := v1.Group("/test")
		testProtected.Use(middleware.JWTAuth(cfg, rd))
		testProtected.Use(middleware.RBAC("founder", "admin"))
		testProtected.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "ok",
				"data":    c.MustGet(middleware.CurrentUserKey),
			})
		})

		// --- Founder: founder / admin (pending founder 可以 apply/me/update-me/invites/产品/收藏) ---
		founder := v1.Group("/founder")
		founder.Use(middleware.JWTAuth(cfg, rd))
		founder.Use(middleware.RBAC("founder", "admin"))
		{
			founder.POST("/invites", func(c *gin.Context) { inviteHdl.FounderGenerate(c) })
			founder.GET("/invites", func(c *gin.Context) { inviteHdl.FounderList(c) })
			founder.POST("/invites/revoke", func(c *gin.Context) { inviteHdl.FounderRevoke(c) })

			founder.POST("/profiles/apply", func(c *gin.Context) { founderHdl.Apply(c) })
			founder.GET("/profiles/me", func(c *gin.Context) { founderHdl.GetMe(c) })
			founder.PUT("/profiles/me", func(c *gin.Context) { founderHdl.UpdateMe(c) })

			founder.GET("/products", func(c *gin.Context) { productHdl.ListMine(c) })
			founder.GET("/products/:id", func(c *gin.Context) { productHdl.GetMine(c) })
			founder.POST("/products", func(c *gin.Context) { productHdl.Create(c) })
			founder.PUT("/products/:id", func(c *gin.Context) { productHdl.Update(c) })
			founder.DELETE("/products/:id", func(c *gin.Context) { productHdl.Delete(c) })
			founder.POST("/products/reorder", func(c *gin.Context) { productHdl.Reorder(c) })

			founder.POST("/favorites/toggle", func(c *gin.Context) { favHdl.Toggle(c) })
			founder.GET("/favorites", func(c *gin.Context) { favHdl.List(c) })
			founder.GET("/favorites/exists/:target_type/:target_id", func(c *gin.Context) { favHdl.Exists(c) })
		}

		// --- Guest/Founder/Admin 共用：任何已登录用户（guest 角色）可收藏、查看自己的收藏、提交举报等 ---
		user := v1.Group("/user")
		user.Use(middleware.JWTAuth(cfg, rd))
		user.Use(middleware.RBAC("guest", "founder", "admin"))
		{
			user.POST("/favorites/toggle", func(c *gin.Context) { favHdl.Toggle(c) })
			user.GET("/favorites", func(c *gin.Context) { favHdl.List(c) })
			user.GET("/favorites/exists/:target_type/:target_id", func(c *gin.Context) { favHdl.Exists(c) })
		}

		// --- Admin: 仅 admin ---
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg, rd))
		admin.Use(middleware.RBAC("admin"))
		{
			admin.GET("/invites", func(c *gin.Context) { inviteHdl.AdminList(c) })
			admin.POST("/invites/generate", func(c *gin.Context) { inviteHdl.AdminGenerateMany(c) })
			admin.POST("/invites/revoke", func(c *gin.Context) { inviteHdl.AdminRevokeMany(c) })

			// Tags CRUD
			admin.GET("/tags", func(c *gin.Context) { tagHdl.AdminListAll(c) })
			admin.POST("/tags", func(c *gin.Context) { tagHdl.AdminCreate(c) })
			admin.PUT("/tags/:id", func(c *gin.Context) { tagHdl.AdminUpdate(c) })
			admin.POST("/tags/:id/deactivate", func(c *gin.Context) { tagHdl.AdminDeactivate(c) })

			// Reports
			admin.GET("/reports", func(c *gin.Context) { reportHdl.AdminList(c) })
			admin.POST("/reports/:id/handle", func(c *gin.Context) { reportHdl.AdminHandle(c) })

			// Founders
			admin.GET("/founders", func(c *gin.Context) {
				page, pageSize := queryPage(c)
				sort := c.Query("sort")
				kw := strings.TrimSpace(c.Query("keyword"))
				var items []*entity.FounderProfile
				var total int64
				var err error
				q := db.WithContext(c.Request.Context()).Model(&entity.FounderProfile{})
				if kw != "" {
					var userIDs []uint64
					_ = db.WithContext(c.Request.Context()).Table("users u").Select("u.id").
						Where("u.username LIKE ? OR u.email LIKE ?", "%"+kw+"%", "%"+kw+"%").
						Scan(&userIDs).Error
					kwLike := "%" + kw + "%"
					q = q.Where("user_id IN ? OR display_name LIKE ? OR slug LIKE ? OR bio LIKE ? OR location LIKE ? OR title LIKE ?",
						userIDs, kwLike, kwLike, kwLike, kwLike, kwLike)
				}
				if e := q.Count(&total).Error; e != nil {
					err = e
				} else {
					switch sort {
					case "favorite_count":
						q = q.Order("favorite_count DESC, id DESC")
					case "view_count":
						q = q.Order("view_count DESC, id DESC")
					default:
						q = q.Order("created_at DESC, id DESC")
					}
					offset := (page - 1) * pageSize
					err = q.Offset(offset).Limit(pageSize).Find(&items).Error
				}
				if err != nil {
					response.Fail(c, http.StatusInternalServerError, "db failed")
					return
				}
				vos := make([]any, 0, len(items))
				for _, fp := range items {
					usr, _ := userRepo.GetByID(c.Request.Context(), fp.UserID)
					username := ""
					email := ""
					var status *entity.UserStatus
					if usr != nil {
						username = usr.Username
						email = usr.Email
						status = &usr.Status
					}
					vos = append(vos, gin.H{
						"id": fp.ID, "user_id": fp.UserID, "username": username, "email": email, "user_status": status,
						"display_name": fp.DisplayName, "title": nsVal(fp.Title), "bio": nsVal(fp.Bio),
						"slug": fp.Slug, "location": nsVal(fp.Location), "avatar_url": nsVal(fp.AvatarURL),
						"cover_url": nsVal(fp.CoverURL), "is_visible": fp.IsVisible,
						"view_count": fp.ViewCount, "favorite_count": fp.FavoriteCount,
						"created_at": fp.CreatedAt, "updated_at": fp.UpdatedAt,
					})
				}
				response.OK(c, gin.H{
					"page":  gin.H{"total": total, "page": page, "size": pageSize},
					"items": vos, "list": vos, "total": total,
				})
			})
			admin.PUT("/founders/:id", func(c *gin.Context) {
				id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
				if id == 0 {
					response.Fail(c, http.StatusBadRequest, "invalid id")
					return
				}
				var body struct {
					IsVisible *bool   `json:"is_visible"`
					Location  *string `json:"location"`
					Title     *string `json:"title"`
					Bio       *string `json:"bio"`
				}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Fail(c, http.StatusBadRequest, "invalid body")
					return
				}
				fields := map[string]any{}
				if body.IsVisible != nil {
					fields["is_visible"] = *body.IsVisible
				}
				if body.Location != nil {
					fields["location"] = *body.Location
				}
				if body.Title != nil {
					fields["title"] = *body.Title
				}
				if body.Bio != nil {
					fields["bio"] = *body.Bio
				}
				if len(fields) > 0 {
					if err := db.WithContext(c.Request.Context()).Model(&entity.FounderProfile{}).Where("id = ?", id).Updates(fields).Error; err != nil {
						response.Fail(c, http.StatusInternalServerError, "db update failed")
						return
					}
					_ = cache.Invalidate(c.Request.Context(), rd, "fp:id:"+strconv.FormatUint(id, 10))
				}
				after, _ := fpRepo.GetByID(c.Request.Context(), id)
				usr, _ := userRepo.GetByID(c.Request.Context(), after.UserID)
				username := ""
				email := ""
				if usr != nil {
					username = usr.Username
					email = usr.Email
				}
				response.OK(c, gin.H{
					"id": after.ID, "user_id": after.UserID, "username": username, "email": email,
					"display_name": after.DisplayName, "title": nsVal(after.Title), "bio": nsVal(after.Bio),
					"slug": after.Slug, "location": nsVal(after.Location), "avatar_url": nsVal(after.AvatarURL),
					"is_visible": after.IsVisible, "view_count": after.ViewCount, "favorite_count": after.FavoriteCount,
				})
			})

			// Products
			admin.GET("/products", func(c *gin.Context) {
				page, pageSize := queryPage(c)
				sort := c.Query("sort")
				kw := strings.TrimSpace(c.Query("keyword"))
				var items []*entity.Product
				var total int64
				var err error
				q := db.WithContext(c.Request.Context()).Model(&entity.Product{})
				if vis := c.Query("is_visible"); vis != "" {
					if vis == "1" || vis == "true" {
						q = q.Where("is_visible = ?", true)
					} else {
						q = q.Where("is_visible = ?", false)
					}
				}
				if c.Query("include_deleted") == "1" || c.Query("include_deleted") == "true" {
					q = q.Unscoped()
				}
				if kw != "" {
					q = q.Where("name LIKE ? OR slug LIKE ? OR one_liner LIKE ?", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
				}
				if e := q.Count(&total).Error; e != nil {
					err = e
				} else {
					switch sort {
					case "favorite_count":
						q = q.Order("favorite_count DESC, id DESC")
					case "view_count":
						q = q.Order("view_count DESC, id DESC")
					default:
						q = q.Order("created_at DESC, id DESC")
					}
					offset := (page - 1) * pageSize
					err = q.Offset(offset).Limit(pageSize).Find(&items).Error
				}
				if err != nil {
					response.Fail(c, http.StatusInternalServerError, "db failed")
					return
				}
				vos := make([]any, 0, len(items))
				for _, p := range items {
					vo, e := productSvc.GetByID(c.Request.Context(), p.ID, true)
					if e == nil && vo != nil {
						vos = append(vos, vo)
					} else {
						vos = append(vos, gin.H{
							"id": p.ID, "slug": p.Slug, "name": p.Name, "tagline": p.Tagline,
							"is_visible": p.IsVisible, "view_count": p.ViewCount, "favorite_count": p.FavoriteCount,
							"created_at": p.CreatedAt, "updated_at": p.UpdatedAt, "deleted_at": p.DeletedAt,
						})
					}
				}
				response.OK(c, gin.H{
					"page":  gin.H{"total": total, "page": page, "size": pageSize},
					"items": vos, "list": vos, "total": total,
				})
			})
			admin.PUT("/products/:id/status", func(c *gin.Context) {
				id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
				if id == 0 {
					response.Fail(c, http.StatusBadRequest, "invalid id")
					return
				}
				var body struct {
					IsVisible *bool  `json:"is_visible"`
					Action    string `json:"action"`
				}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Fail(c, http.StatusBadRequest, "invalid body")
					return
				}
				if body.Action == "soft_delete" {
					if err := db.WithContext(c.Request.Context()).Delete(&entity.Product{}, id).Error; err != nil {
						response.Fail(c, http.StatusInternalServerError, "delete failed")
						return
					}
				} else if body.Action == "restore" {
					if err := db.WithContext(c.Request.Context()).Unscoped().Model(&entity.Product{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
						response.Fail(c, http.StatusInternalServerError, "restore failed")
						return
					}
				}
				if body.IsVisible != nil {
					_ = db.WithContext(c.Request.Context()).Unscoped().Model(&entity.Product{}).Where("id = ?", id).Update("is_visible", *body.IsVisible).Error
				}
				vo, _ := productSvc.GetByID(c.Request.Context(), id, true)
				if vo == nil {
					var p entity.Product
					_ = db.WithContext(c.Request.Context()).Unscoped().Where("id = ?", id).Take(&p).Error
					response.OK(c, gin.H{
						"id": p.ID, "slug": p.Slug, "name": p.Name, "is_visible": p.IsVisible,
						"view_count": p.ViewCount, "favorite_count": p.FavoriteCount,
					})
					return
				}
				response.OK(c, vo)
			})

			// Users
			admin.GET("/users", func(c *gin.Context) {
				page, pageSize := queryPage(c)
				role := strings.TrimSpace(c.Query("role"))
				status := strings.TrimSpace(c.Query("status"))
				kw := strings.TrimSpace(c.Query("keyword"))
				q := db.WithContext(c.Request.Context()).Model(&entity.User{}).Unscoped()
				if role != "" {
					q = q.Where("role = ?", role)
				}
				if status != "" {
					q = q.Where("status = ?", status)
				}
				if kw != "" {
					q = q.Where("username LIKE ? OR email LIKE ?", "%"+kw+"%", "%"+kw+"%")
				}
				var total int64
				var items []*entity.User
				var err error
				if e := q.Count(&total).Error; e != nil {
					err = e
				} else {
					offset := (page - 1) * pageSize
					err = q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error
				}
				if err != nil {
					response.Fail(c, http.StatusInternalServerError, "db failed")
					return
				}
				list := make([]any, 0, len(items))
				for _, u := range items {
					fp, _ := fpRepo.GetByUserID(c.Request.Context(), u.ID)
					founderSlug := ""
					if fp != nil {
						founderSlug = fp.Slug
					}
					list = append(list, gin.H{
						"id": u.ID, "username": u.Username, "email": u.Email,
						"role": u.Role, "status": u.Status, "avatar_url": nsVal(u.AvatarURL),
						"register_source": u.RegisterSource, "last_login_at": u.LastLoginAt,
						"created_at": u.CreatedAt, "updated_at": u.UpdatedAt, "deleted_at": u.DeletedAt,
						"founder_slug": founderSlug,
					})
				}
				response.OK(c, gin.H{"page": gin.H{"total": total, "page": page, "size": pageSize}, "items": list, "list": list})
			})
			admin.PUT("/users/:id", func(c *gin.Context) {
				id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
				if id == 0 {
					response.Fail(c, http.StatusBadRequest, "invalid id")
					return
				}
				var body struct {
					Role   *string            `json:"role"`
					Status *entity.UserStatus `json:"status"`
				}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Fail(c, http.StatusBadRequest, "invalid body")
					return
				}
				fields := map[string]any{}
				if body.Role != nil {
					fields["role"] = *body.Role
				}
				if body.Status != nil {
					fields["status"] = *body.Status
				}
				if len(fields) > 0 {
					if err := db.WithContext(c.Request.Context()).Model(&entity.User{}).Where("id = ?", id).Updates(fields).Error; err != nil {
						response.Fail(c, http.StatusInternalServerError, "db update failed")
						return
					}
				}
				u, _ := userRepo.GetByID(c.Request.Context(), id)
				response.OK(c, u)
			})
		}
	}
	slog.Info("routes registered")
	return r
}

func runMigrations(cfg *config.Config, up bool, downN int) error {
	return migratepkg.Run(cfg, up, downN)
}

func queryPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func nsVal(ns *string) string {
	if ns == nil {
		return ""
	}
	return *ns
}
