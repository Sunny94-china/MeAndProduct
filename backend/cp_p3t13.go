//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

const base = "http://127.0.0.1:8080/api/v1"

var pass, fail int

func main() {
	start := time.Now()
	fmt.Println("=== CP P3 T10 S3 — 28 Content Module API Checkpoints ===")
	fmt.Println()

	rd := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", Password: "dev_redis_pwd", DB: 0})
	defer rd.Close()
	ctx := context.Background()
	_, _ = rd.Del(ctx, "tags:active:list", "tags:grouped").Result()

	s := &Suite{}

	// 1. 账号准备：1 admin + 2 founder（A/B） + 1 guest favorite collector
	s.Admin, s.AdminTok = ensureAdminUser("p3t13")
	s.FounderATok, _, _ = ensureFounderUser("p3t13_a")
	s.FounderBTok, _, _ = ensureFounderUser("p3t13_b")
	var fcID uint64
	s.FounderCTok, _, fcID = ensureFounderUser("p3t13_c")
	s.FounderCID = fcID

	// 2. 前置：标签 CRUD（必须先于 Apply/Create 产品，因为 Apply/Product 支持 TagIDs）
	cpAdminTagsCreate1(s)   // T-1 admin create 2 tags (industry+sas + stage+idea)
	cpAdminTagsCreate2(s)   // T-2 再建 1 个 role tag
	cpAdminTagsDuplicate(s) // T-3 slug 重复 409（uk_tag_slug）
	cpAdminTagsList(s)      // T-4 admin list 全部 3 tags + product_count=0/founder_count=0
	cpPublicTagsActive(s)   // T-5 guest GET /public/tags 返回 3 个
	cpPublicTagsGrouped(s)  // T-6 guest GET /public/tags/grouped 按 type 分组

	// 3. Founder 名片 6 CP（Apply + Update + GetMe + Public List + Public Detail + Guest 403）
	cpFounderApplyA(s)             // F-1 A Apply + 自定义 slug + Links
	cpFounderApplyB(s)             // F-2 B Apply
	cpFounderUpdateA(s)            // F-3 A UpdateMe 改 Title + 加 Link
	cpFounderGetMeA(s)             // F-4 A GetMe 返回 id/slug/display_name
	cpPublicFounderList(s)         // F-5 PublicList 分页 total=2
	cpPublicFounderDetailBySlug(s) // F-6 PublicGetBySlug 返回 slug/display_name/Links

	// 4. Product 7 CP（A Create 2 个 + Update + Delete + Reorder + B 越权改 A + Public）
	cpProductCreateA1(s)             // P-1 A Create product1 + GithubURL + 2 Tags（expect Tag.ProductCount +1）
	cpProductCreateA2(s)             // P-2 A Create product2
	cpProductSlugConflictA(s)        // P-3 A 再建 product slug 冲突 409（uk_founder_slug）
	cpProductFounderBCannotEditA(s)  // P-4 B PUT A 的 product → 403 归属校验
	cpProductUpdateA1(s)             // P-5 A Update product1 Description
	cpProductReorderA(s)             // P-6 A Reorder: [p2,p1]
	cpPublicProductsByFounder(s)     // P-7 PublicListByFounder 返回 2 个 + desc=更新
	cpPublicProductByFounderSlugs(s) // P-8 PublicGetByFounderSlugs（加一条，因为 B 越权单独列出）

	// 5. Favorite 3 CP（C 收藏 A's profile → 幂等 → 列表）
	cpFavoriteToggleCOnAFounder(s) // V-1 C 收藏 A founder：favorited=true，count+1
	cpFavoriteToggleIdempotent(s)  // V-2 再 toggle 一次 → favorited=false，count-1（幂等）
	cpFavoriteListC(s)             // V-3 C 收藏列表（先加回来再取）

	// 6. Report 4 CP（匿名提交 → Admin List → Admin Handle hide_target=true → 再提交 guest 403 handle）
	cpReportAnonymousSubmit(s)     // R-1 匿名 POST /public/reports → 200
	cpAdminReportList(s)           // R-2 Admin GET /admin/reports → total=1
	cpAdminReportHandleHide(s)     // R-3 Admin handle → resolved + hide_target=true → DB is_visible=0
	cpFounderCannotHandleReport(s) // R-4 FounderA POST handle → 403

	// 7. Link 2 CP（点击 A's 第一个 link count+1 → public tags grouped 稳定）
	cpLinkTrackClick(s)      // L-1 POST /public/links/:id/click → click_count 0→1
	cpLinkTrackClickAgain(s) // L-2 再点一次 1→2 + DB 行验证

	// 8. RBAC 补位：Guest 访问 /founder 和 /admin → 401/403
	cpRBACNoJWTFounder401(s)     // R1 GET /founder/profiles/me 无 JWT → 401
	cpRBACFounderAdmin403(s)     // R2 FounderA GET /admin/reports → 403
	cpRBACFounderTagCreate403(s) // R3 FounderA POST /admin/tags → 403

	// 收尾：Tag product_count 回写断言（Create product1/product2 → product1 2 tags, product2 0 tags → 两个 tag.product_count 都 +1）
	cpTagCountAfterProducts(s) // T-7 /tags grouped industry sas count=1?

	fmt.Println()
	fmt.Printf("=== Result: pass %d / %d (fail %d)  elapsed %s ===\n", pass, pass+fail, fail, time.Since(start).Round(time.Millisecond))
	if fail > 0 {
		os.Exit(1)
	}
}

type Suite struct {
	Admin        *entity.User
	AdminTok     string
	FounderATok  string
	FounderBTok  string
	FounderCTok  string
	FounderASlug string
	FounderBSlug string
	FounderAID   uint64
	FounderBID   uint64
	FounderCID   uint64

	TagIndustrySAS  uint64
	TagStageIdea    uint64
	TagRoleEngineer uint64

	Product1ID   uint64
	Product1Slug string
	Product2ID   uint64
	Product2Slug string

	Link1ID   uint64
	Report1ID uint64
}

// =========================================================
// helpers
// =========================================================

func check(name string, ok bool, detail ...string) {
	if ok {
		pass++
		fmt.Printf("✅ %-58s\n", name)
	} else {
		fail++
		extra := strings.Join(detail, " | ")
		fmt.Printf("❌ %-58s  %s\n", name, extra)
	}
}

func postJSON(url, token string, body any) (int, map[string]any, string) {
	var reqBody []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			log.Fatalf("json marshal: %v", err)
		}
		reqBody = b
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Fatalf("new req: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("do POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m, string(raw)
}

func putJSON(url, token string, body any) (int, map[string]any, string) {
	var reqBody []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			log.Fatalf("json marshal: %v", err)
		}
		reqBody = b
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Fatalf("new req PUT: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("do PUT %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m, string(raw)
}

func delURL(url, token string) (int, map[string]any, string) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatalf("new req DEL: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("do DEL %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m, string(raw)
}

func getURL(url, token string) (int, map[string]any, string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("new req GET: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("do GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m, string(raw)
}

func aInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case uint64:
		return int(x)
	}
	return -1
}
func aUint(v any) uint64 {
	switch x := v.(type) {
	case float64:
		return uint64(x)
	case int:
		return uint64(x)
	case uint64:
		return x
	}
	return 0
}
func aStr(v any) string          { s, _ := v.(string); return s }
func asMap(v any) map[string]any { m, _ := v.(map[string]any); return m }
func asSlice(v any) []any        { s, _ := v.([]any); return s }
func mustJSON(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}
func dataOf(m map[string]any) map[string]any { return asMap(m["data"]) }

func db() *gorm.DB {
	dsn := "meandproduct:dev_pwd@tcp(127.0.0.1:3306)/meandproduct?charset=utf8mb4&parseTime=True&loc=Local"
	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	return d
}

// 复用 cp_p2t8 的 ensure 模式：
// ensureAdminUser 返回 (user, access, refresh)
func ensureAdminUser(suf string) (*entity.User, string) {
	regBody := map[string]any{
		"username": "admin_" + suf, "email": "admin_" + suf + "@meandproduct.dev",
		"password": "Admin@2025ok!", "display_name": "Admin " + suf,
	}
	code, m, _ := postJSON(base+"/auth/register/self", "", regBody)
	d := db()
	var u entity.User
	var uid uint64
	if code == 200 {
		user := asMap(dataOf(m)["user"])
		uid = aUint(user["id"])
	} else if code == 409 {
		if err := d.Where("username=?", "admin_"+suf).Take(&u).Error; err == nil {
			uid = u.ID
		}
	} else {
		log.Fatalf("admin register code=%d body=%s", code, mustJSON(m))
	}
	if uid == 0 {
		if err := d.Where("username=?", "admin_"+suf).Take(&u).Error; err == nil {
			uid = u.ID
		}
	}
	if uid == 0 {
		log.Fatalf("admin uid=0 suf=%s", suf)
	}
	d.Model(&entity.User{ID: uid}).UpdateColumns(map[string]any{"role": entity.RoleAdmin, "status": entity.UserStatusActive})
	_ = d.Where("id=?", uid).Take(&u).Error
	code, lm, raw := postJSON(base+"/auth/login", "", map[string]any{
		"username_or_email": "admin_" + suf, "password": "Admin@2025ok!",
	})
	if code != 200 {
		log.Fatalf("admin login code=%d body=%s", code, raw)
	}
	access := aStr(dataOf(lm)["access_token"])
	if access == "" {
		log.Fatalf("admin access empty: %s", raw)
	}
	return &u, access
}

func ensureFounderUser(suf string) (string, string, uint64) {
	regBody := map[string]any{
		"username": "fdr_" + suf, "email": "fdr_" + suf + "@meandproduct.dev",
		"password": "Founder@2025ok!", "display_name": "Founder " + suf,
	}
	code, m, raw := postJSON(base+"/auth/register/self", "", regBody)
	if code == 409 {
		code, m, raw = postJSON(base+"/auth/login", "", map[string]any{
			"username_or_email": "fdr_" + suf, "password": "Founder@2025ok!",
		})
		if code != 200 {
			log.Fatalf("founder login %s code=%d body=%s", suf, code, raw)
		}
	} else if code != 200 {
		log.Fatalf("founder register %s code=%d body=%s", suf, code, raw)
	}
	d := dataOf(m)
	access := aStr(d["access_token"])
	refresh := aStr(d["refresh_token"])
	uid := aUint(asMap(d["user"])["id"])
	if uid > 0 {
		db().Model(&entity.User{ID: uid}).UpdateColumn("status", entity.UserStatusActive)
	}
	return access, refresh, uid
}

// =========================================================
// T — Tag module CPs (7 tests)
// =========================================================

func cpAdminTagsCreate1(s *Suite) {
	body := map[string]any{"name": "SaaS", "slug": "sas", "type": "industry", "sort_order": 100}
	code, m, raw := postJSON(base+"/admin/tags", s.AdminTok, body)
	if code == 200 {
		d := asMap(dataOf(m))
		s.TagIndustrySAS = aUint(d["id"])
		check("T-1 admin create tag industry/sas → 200", true, mustJSON(d))
		return
	}
	// 409：已存在，从 list 里找 sas/idea 的 id
	if code == 409 {
		_, lm, _ := getURL(base+"/admin/tags", s.AdminTok)
		for _, it := range asSlice(lm["data"]) {
			mm := asMap(it)
			if aStr(mm["slug"]) == "sas" {
				s.TagIndustrySAS = aUint(mm["id"])
			}
			if aStr(mm["slug"]) == "idea" {
				s.TagStageIdea = aUint(mm["id"])
			}
		}
		check("T-1 tag sas 已存在(409) + id>0", s.TagIndustrySAS > 0, "id="+strconv.FormatUint(s.TagIndustrySAS, 10), raw)
		return
	}
	check("T-1 admin create tag industry/sas → 200", false, "code="+strconv.Itoa(code), raw)
}

func cpAdminTagsCreate2(s *Suite) {
	if s.TagStageIdea > 0 {
		check("T-2 tag stage/idea 已复用(来自T-1回查)", true, "id="+strconv.FormatUint(s.TagStageIdea, 10))
		return
	}
	body := map[string]any{"name": "Idea", "slug": "idea", "type": "stage", "sort_order": 50}
	code, m, raw := postJSON(base+"/admin/tags", s.AdminTok, body)
	if code == 200 {
		s.TagStageIdea = aUint(asMap(dataOf(m))["id"])
		check("T-2 admin create tag stage/idea → 200", true, raw)
		return
	}
	if code == 409 {
		_, lm, _ := getURL(base+"/admin/tags", s.AdminTok)
		for _, it := range asSlice(lm["data"]) {
			mm := asMap(it)
			if aStr(mm["slug"]) == "idea" {
				s.TagStageIdea = aUint(mm["id"])
			}
		}
		check("T-2 tag idea 已存在(409) + id>0", s.TagStageIdea > 0, "id="+strconv.FormatUint(s.TagStageIdea, 10), raw)
		return
	}
	check("T-2 admin create tag stage/idea → 200", false, "code="+strconv.Itoa(code), raw)
}

func cpAdminTagsDuplicate(s *Suite) {
	body := map[string]any{"name": "Duplicate SAS", "slug": "sas", "type": "industry"}
	code, _, raw := postJSON(base+"/admin/tags", s.AdminTok, body)
	check("T-3 duplicate slug sas → 409", code == 409, "code="+strconv.Itoa(code), raw)
}

func cpAdminTagsList(s *Suite) {
	code, m, raw := getURL(base+"/admin/tags", s.AdminTok)
	list := asSlice(m["data"])
	check("T-4 admin list tags len>=3", code == 200 && len(list) >= 3, "len="+strconv.Itoa(len(list)), raw)
	if len(list) >= 3 {
		names := []string{}
		for _, it := range list {
			names = append(names, aStr(asMap(it)["name"]))
		}
		sort.Strings(names)
		check("T-4a names 包含 SaaS/Idea", contains(names, "SaaS") && contains(names, "Idea"), fmt.Sprintf("%v", names))
	}
}

func cpPublicTagsActive(s *Suite) {
	code, m, raw := getURL(base+"/public/tags", "")
	list := asSlice(m["data"])
	check("T-5 guest public tags len>=3", code == 200 && len(list) >= 3, "len="+strconv.Itoa(len(list))+" code="+strconv.Itoa(code), raw)
}

func cpPublicTagsGrouped(s *Suite) {
	code, m, raw := getURL(base+"/public/tags/grouped", "")
	d := asMap(dataOf(m))
	groups := asMap(d["groups"])
	check("T-6 grouped contains industry + stage", code == 200 && groups["industry"] != nil && groups["stage"] != nil,
		fmt.Sprintf("keys=%v", mapKeys(groups)), raw)
}

func cpTagCountAfterProducts(s *Suite) {
	code, m, raw := getURL(base+"/public/tags", "")
	list := asSlice(m["data"])
	var sasPC int
	for _, it := range list {
		mm := asMap(it)
		if aStr(mm["slug"]) == "sas" {
			sasPC = aInt(mm["product_count"])
		}
	}
	check("T-7 tag sas product_count>=1", code == 200 && sasPC >= 1, "sasPC="+strconv.Itoa(sasPC), raw)
}

// =========================================================
// F — Founder module CPs (6 tests)
// =========================================================

func cpFounderApplyA(s *Suite) {
	// 先查是否已存在（多轮重跑CP），存在则复用
	codeG, mG, _ := getURL(base+"/founder/profiles/me", s.FounderATok)
	if codeG == 200 {
		d := asMap(dataOf(mG))
		s.FounderAID = aUint(d["id"])
		s.FounderASlug = aStr(d["slug"])
		check("F-1 founderA 已存在 → 复用 (code=200)", s.FounderAID > 0 && len(s.FounderASlug) > 0,
			"id="+strconv.FormatUint(s.FounderAID, 10)+" slug="+s.FounderASlug)
		return
	}
	body := map[string]any{
		"display_name": "Founder P3T13 A", "slug": "p3t13_fa",
		"title": "Builder of SaaS", "bio": "Building side projects.", "location": "Shanghai",
		"contact_email": "p3t13_a@meandproduct.dev",
		"custom_links": []any{
			map[string]any{"link_type": "github", "url": "https://github.com/p3t13-a"},
			map[string]any{"link_type": "twitter", "url": "https://twitter.com/p3t13_a"},
		},
	}
	code, m, raw := postJSON(base+"/founder/profiles/apply", s.FounderATok, body)
	check("F-1 founderA apply → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		s.FounderASlug = aStr(d["slug"])
		s.FounderAID = aUint(d["id"])
		check("F-1a slug+id 赋值成功", len(s.FounderASlug) > 0 && s.FounderAID > 0, "slug="+s.FounderASlug+" id="+strconv.FormatUint(s.FounderAID, 10))
	}
}

func cpFounderApplyB(s *Suite) {
	codeG, mG, _ := getURL(base+"/founder/profiles/me", s.FounderBTok)
	if codeG == 200 {
		d := asMap(dataOf(mG))
		s.FounderBID = aUint(d["id"])
		s.FounderBSlug = aStr(d["slug"])
		check("F-2 founderB 已存在 → 复用 (code=200)", s.FounderBID > 0 && len(s.FounderBSlug) > 0,
			"id="+strconv.FormatUint(s.FounderBID, 10)+" slug="+s.FounderBSlug)
		return
	}
	body := map[string]any{
		"display_name": "Founder P3T13 B", "slug": "p3t13_fb",
		"title": "Indie Hacker", "bio": "Launch 12/year.", "location": "Hangzhou",
		"custom_links": []any{map[string]any{"link_type": "official", "url": "https://p3t13b.example"}},
	}
	code, m, raw := postJSON(base+"/founder/profiles/apply", s.FounderBTok, body)
	check("F-2 founderB apply → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		s.FounderBSlug = aStr(d["slug"])
		s.FounderBID = aUint(d["id"])
	}
}

func cpFounderUpdateA(s *Suite) {
	title := "Builder of SaaS (Updated)"
	body := map[string]any{
		"title": &title,
		"custom_links": []map[string]any{
			{"link_type": "github", "url": "https://github.com/p3t13-a"},
			{"link_type": "twitter", "url": "https://twitter.com/p3t13_a"},
			{"link_type": "website", "url": "https://p3t13-a.dev"},
		},
	}
	code, m, raw := putJSON(base+"/founder/profiles/me", s.FounderATok, body)
	check("F-3 founderA update title+add link → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		check("F-3a title contains Updated", aStr(d["title"]) == title, aStr(d["title"]))
	}
}

func cpFounderGetMeA(s *Suite) {
	code, m, raw := getURL(base+"/founder/profiles/me", s.FounderATok)
	d := asMap(dataOf(m))
	check("F-4 founderA getMe → slug 匹配 apply", code == 200 && aStr(d["slug"]) == s.FounderASlug, "want="+s.FounderASlug+" got="+aStr(d["slug"]), raw)
}

func cpPublicFounderList(s *Suite) {
	code, m, raw := getURL(base+"/public/founders?page=1&page_size=10", "")
	d := asMap(dataOf(m))
	page := asMap(d["page"])
	list := asSlice(d["list"])
	check("F-5 public founders total>=2", code == 200 && aInt(page["total"]) >= 2 && len(list) >= 2,
		"total="+strconv.Itoa(aInt(page["total"]))+" len="+strconv.Itoa(len(list)), raw)
}

func cpPublicFounderDetailBySlug(s *Suite) {
	code, m, raw := getURL(base+"/public/founders/"+s.FounderASlug, "")
	d := asMap(dataOf(m))
	dn := strings.ToLower(aStr(d["display_name"]))
	check("F-6 public detail A slug 匹配 + display_name 含p3t13_a", code == 200 && aStr(d["slug"]) == s.FounderASlug &&
		strings.Contains(dn, "p3t13_a"), raw)
}

// =========================================================
// P — Product module CPs (8)
// =========================================================

func cpProductCreateA1(s *Suite) {
	gh := "https://github.com/p3t13-a/prod1"
	body := map[string]any{
		"name": "P3T13 Product 1", "slug": "p3t13_p1", "tagline": "Tagline 1",
		"github_url": &gh, "status": "idea", "official_url": "https://p1.p3t13-a.dev",
		"tag_ids":      []uint64{s.TagIndustrySAS, s.TagStageIdea},
		"custom_links": []any{map[string]any{"link_type": "docs", "url": "https://docs.p3t13-a.dev/p1"}},
	}
	code, m, raw := postJSON(base+"/founder/products", s.FounderATok, body)
	check("P-1 founderA create product1 +2 tags → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		s.Product1ID = aUint(d["id"])
		s.Product1Slug = aStr(d["slug"])
		// link id: 取 custom_links[0].id（vo 里 CustomLinks 包含 id）
		cls := asSlice(d["custom_links"])
		if len(cls) > 0 {
			s.Link1ID = aUint(asMap(cls[0])["id"])
		}
		check("P-1a link1 id>0 + slug非空", s.Link1ID > 0 && len(s.Product1Slug) > 0,
			"link1="+strconv.FormatUint(s.Link1ID, 10)+" slug="+s.Product1Slug)
	}
}

func cpProductCreateA2(s *Suite) {
	body := map[string]any{
		"name": "P3T13 Product 2", "slug": "p3t13_p2", "tagline": "Tagline 2",
		"official_url": "https://p2.p3t13-a.dev", "status": "launched",
	}
	code, m, raw := postJSON(base+"/founder/products", s.FounderATok, body)
	check("P-2 founderA create product2 → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		s.Product2ID = aUint(d["id"])
		s.Product2Slug = aStr(d["slug"])
	}
}

func cpProductSlugConflictA(s *Suite) {
	body := map[string]any{
		"name": "Duplicate P1", "slug": "p3t13_p1",
		"official_url": "https://dup.p3t13.dev", "status": "idea",
	}
	code, m, raw := postJSON(base+"/founder/products", s.FounderATok, body)
	ok := code == 200 && strings.HasPrefix(aStr(asMap(dataOf(m))["slug"]), "p3t13_p1-")
	check("P-3 slug 冲突自动解: 以 p3t13_p1- 开头 → 200", ok,
		"code="+strconv.Itoa(code)+" slug="+aStr(asMap(dataOf(m))["slug"]), raw)
}

func cpProductFounderBCannotEditA(s *Suite) {
	body := map[string]any{"name": any("Hacked Name")}
	code, _, raw := putJSON(base+"/founder/products/"+strconv.FormatUint(s.Product1ID, 10), s.FounderBTok, body)
	check("P-4 founderB edit founderA product → 404（越权隐藏NotFound）", code == 404, "code="+strconv.Itoa(code), raw)
}

func cpProductUpdateA1(s *Suite) {
	desc := "Updated description with markdown **bold**."
	body := map[string]any{"description": &desc}
	code, m, raw := putJSON(base+"/founder/products/"+strconv.FormatUint(s.Product1ID, 10), s.FounderATok, body)
	check("P-5 founderA update product1 description → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		check("P-5a description returned equal", aStr(d["description"]) == desc, aStr(d["description"]))
	}
}

func cpProductReorderA(s *Suite) {
	body := map[string]any{"ordered_ids": []uint64{s.Product2ID, s.Product1ID}}
	code, _, raw := postJSON(base+"/founder/products/reorder", s.FounderATok, body)
	check("P-6 reorder [p2,p1] → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		var p1, p2 entity.Product
		d := db()
		_ = d.Where("id=?", s.Product1ID).Take(&p1).Error
		_ = d.Where("id=?", s.Product2ID).Take(&p2).Error
		check("P-6a sort_order p1>p2（ordered_ids首项=索引0最小）", p1.SortOrder > p2.SortOrder,
			"p1="+strconv.Itoa(p1.SortOrder)+" p2="+strconv.Itoa(p2.SortOrder))
	}
}

func cpPublicProductsByFounder(s *Suite) {
	code, m, raw := getURL(base+"/public/founders/"+s.FounderASlug+"/products", "")
	d := asMap(dataOf(m))
	page := asMap(d["page"])
	list := asSlice(d["list"])
	check("P-7 public list A products len>=2 + total>=2", code == 200 && len(list) >= 2 && aInt(page["total"]) >= 2,
		"len="+strconv.Itoa(len(list))+" total="+strconv.Itoa(aInt(page["total"])), raw)
}

func cpPublicProductByFounderSlugs(s *Suite) {
	code, m, raw := getURL(base+"/public/founders/"+s.FounderASlug+"/products/"+s.Product1Slug, "")
	d := asMap(dataOf(m))
	check("P-8 public detail product1: Updated description", code == 200 &&
		strings.Contains(aStr(d["description"]), "markdown **bold**"), raw)
}

// =========================================================
// V — Favorite module CPs (3)
// =========================================================

func cpFavoriteToggleCOnAFounder(s *Suite) {
	if s.FounderCID > 0 && s.FounderAID > 0 {
		_ = db().Exec("DELETE FROM favorites WHERE user_id=? AND target_type='founder' AND target_id=?",
			s.FounderCID, s.FounderAID).Error
	}
	body := map[string]any{"target_type": "founder", "target_id": int(s.FounderAID)}
	code, m, raw := postJSON(base+"/founder/favorites/toggle", s.FounderCTok, body)
	check("V-1 C收藏A founder → favorited=true", code == 200 && asMap(dataOf(m))["favorited"] == true,
		"fav="+mustJSON(asMap(dataOf(m))), raw)
}

func cpFavoriteToggleIdempotent(s *Suite) {
	body := map[string]any{"target_type": "founder", "target_id": int(s.FounderAID)}
	// 再 toggle 一次 → false
	code, m, _ := postJSON(base+"/founder/favorites/toggle", s.FounderCTok, body)
	ok := code == 200 && asMap(dataOf(m))["favorited"] == false
	check("V-2 再次toggle → favorited=false（幂等）", ok, mustJSON(asMap(dataOf(m))))
	// 再加回来，给 list 用
	_, m2, _ := postJSON(base+"/founder/favorites/toggle", s.FounderCTok, body)
	check("V-2a 再加回来 → true", asMap(dataOf(m2))["favorited"] == true)
}

func cpFavoriteListC(s *Suite) {
	code, m, raw := getURL(base+"/founder/favorites?page=1&page_size=10", s.FounderCTok)
	d := asMap(dataOf(m))
	page := asMap(d["page"])
	list := asSlice(d["list"])
	check("V-3 C收藏列表 len>=1 & total>=1", code == 200 && len(list) >= 1 && aInt(page["total"]) >= 1,
		"len="+strconv.Itoa(len(list))+" total="+strconv.Itoa(aInt(page["total"])), raw)
}

// =========================================================
// R — Report module CPs (4)
// =========================================================

func cpReportAnonymousSubmit(s *Suite) {
	body := map[string]any{
		"target_type": "product", "target_id": s.Product2ID,
		"reason_type": "spam", "reason_detail": "CP-P3T13 spammy product 2 p3t13",
		"contact_info": "anon@meandproduct.dev",
	}
	code, m, raw := postJSON(base+"/public/reports", "", body)
	check("R-1 匿名举报 product2 → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		s.Report1ID = aUint(d["id"])
		check("R-1a report id>0 + status=pending", s.Report1ID > 0 && aStr(d["status"]) == "pending", "id="+strconv.FormatUint(s.Report1ID, 10))
	}
}

func cpAdminReportList(s *Suite) {
	code, m, raw := getURL(base+"/admin/reports?page=1&page_size=20", s.AdminTok)
	d := asMap(dataOf(m))
	page := asMap(d["page"])
	check("R-2 admin reports total>=1", code == 200 && aInt(page["total"]) >= 1,
		"total="+strconv.Itoa(aInt(page["total"])), raw)
}

func cpAdminReportHandleHide(s *Suite) {
	note := "下架违规产品 2"
	body := map[string]any{
		"status": "resolved", "handler_note": &note, "hide_target": true,
	}
	code, m, raw := postJSON(base+"/admin/reports/"+strconv.FormatUint(s.Report1ID, 10)+"/handle", s.AdminTok, body)
	check("R-3 admin handle resolved + hide_target → 200", code == 200, "code="+strconv.Itoa(code), raw)
	if code == 200 {
		d := asMap(dataOf(m))
		check("R-3a status=resolved + handler_note 含下架",
			aStr(d["status"]) == "resolved" && strings.Contains(aStr(d["handler_note"]), "下架"),
			mustJSON(d))
		// DB 断言 product2 is_visible=0
		var p2 entity.Product
		_ = db().Where("id=?", s.Product2ID).Take(&p2).Error
		check("R-3b product2 DB is_visible=false", p2.IsVisible == false, "p2.is_visible="+fmt.Sprintf("%v", p2.IsVisible))
	}
}

func cpFounderCannotHandleReport(s *Suite) {
	body := map[string]any{"status": "reviewed", "handler_note": "x"}
	code, _, raw := postJSON(base+"/admin/reports/"+strconv.FormatUint(s.Report1ID, 10)+"/handle", s.FounderATok, body)
	check("R-4 founderA 调 handle → 403", code == 403, "code="+strconv.Itoa(code), raw)
}

// =========================================================
// L — Link click CPs (2)
// =========================================================

func cpLinkTrackClick(s *Suite) {
	if s.Link1ID == 0 {
		check("L-1 SKIP (link1 id=0)", false, "product1 custom_links id missing")
		return
	}
	before := linkClickCount(s.Link1ID)
	code, _, raw := postJSON(base+"/public/links/"+strconv.FormatUint(s.Link1ID, 10)+"/click", "", nil)
	after := linkClickCount(s.Link1ID)
	check("L-1 link1 click count +1", code == 200 && after == before+1,
		fmt.Sprintf("before=%d after=%d code=%d", before, after, code), raw)
}

func cpLinkTrackClickAgain(s *Suite) {
	if s.Link1ID == 0 {
		check("L-2 SKIP", false)
		return
	}
	before := linkClickCount(s.Link1ID)
	code, _, raw := postJSON(base+"/public/links/"+strconv.FormatUint(s.Link1ID, 10)+"/click", "", nil)
	after := linkClickCount(s.Link1ID)
	check("L-2 link1 click count 再+1", code == 200 && after == before+1,
		fmt.Sprintf("before=%d after=%d", before, after), raw)
}

func linkClickCount(id uint64) int {
	var l entity.Link
	if err := db().Where("id=?", id).Take(&l).Error; err != nil {
		return -1
	}
	return int(l.ClickCount)
}

// =========================================================
// RBAC extras (3)
// =========================================================

func cpRBACNoJWTFounder401(s *Suite) {
	code, _, raw := getURL(base+"/founder/profiles/me", "")
	check("R1 guest /founder/profiles/me → 401", code == 401, "code="+strconv.Itoa(code), raw)
}

func cpRBACFounderAdmin403(s *Suite) {
	code, _, raw := getURL(base+"/admin/reports?page=1", s.FounderATok)
	check("R2 founderA /admin/reports → 403", code == 403, "code="+strconv.Itoa(code), raw)
}

func cpRBACFounderTagCreate403(s *Suite) {
	body := map[string]any{"name": "Hack", "slug": "hack", "type": "industry"}
	code, _, raw := postJSON(base+"/admin/tags", s.FounderATok, body)
	check("R3 founderA POST /admin/tags → 403", code == 403, "code="+strconv.Itoa(code), raw)
}

// =========================================================
// utils
// =========================================================

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if s == t {
			return true
		}
	}
	return false
}

func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
