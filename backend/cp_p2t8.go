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
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

var pass, fail int

func check(name string, ok bool, detail ...string) {
	if ok {
		pass++
		fmt.Printf("✅ %s\n", name)
	} else {
		fail++
		extra := strings.Join(detail, " | ")
		fmt.Printf("❌ %s  %s\n", name, extra)
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

func getURL(url, token string) (int, map[string]any, string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("new req: %v", err)
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
func aFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	}
	return -1
}
func aStr(v any) string {
	s, _ := v.(string)
	return s
}
func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}
func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

func db() *gorm.DB {
	dsn := "meandproduct:dev_pwd@tcp(127.0.0.1:3306)/meandproduct?charset=utf8mb4&parseTime=True&loc=Local"
	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	return d
}

func ensureAdminUser(username string) (*entity.User, string, string) {
	// 注册（自助注册默认 founder+pending → DB 直接改 role=admin，status=active）
	suf := username
	regBody := map[string]any{
		"username":     "admin_" + suf,
		"email":        "admin_" + suf + "@meandproduct.dev",
		"password":     "Admin@2025ok!",
		"display_name": "Admin " + suf,
	}
	code, m, _ := postJSON("http://127.0.0.1:8080/api/v1/auth/register/self", "", regBody)
	var uid uint64
	var u entity.User
	// 1. 确保用户存在（新用户 → register 创建；旧用户 → 登录前 DB 改 active）
	d := db()
	if code == 200 {
		d := asMap(m["data"])
		user := asMap(d["user"])
		uid = uint64(aFloat(user["id"]))
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
	// 2. DB 改成 admin + active（必须在重新登录前做！）
	if uid == 0 {
		log.Fatalf("admin user uid 始终为 0, suf=%s", suf)
	}
	d.Model(&entity.User{ID: uid}).UpdateColumns(map[string]any{
		"role":   entity.RoleAdmin,
		"status": entity.UserStatusActive,
	})
	_ = d.Where("id=?", uid).Take(&u).Error
	// 3. 登录拿 **最新 claims（role=admin）** 的新 token（非常关键！否则 JWT 里 role 还是 founder/pending 导致 403）
	loginBody := map[string]any{
		"username_or_email": "admin_" + suf,
		"password":          "Admin@2025ok!",
	}
	loginCode, lm, raw := postJSON("http://127.0.0.1:8080/api/v1/auth/login", "", loginBody)
	if loginCode != 200 {
		log.Fatalf("admin login failed: code=%d body=%s", loginCode, raw)
	}
	data := asMap(lm["data"])
	access := aStr(data["access_token"])
	refresh := aStr(data["refresh_token"])
	if access == "" {
		log.Fatalf("admin login access_token empty body=%s", raw)
	}
	return &u, access, refresh
}

func ensureFounderUser(suf string) (string, string) {
	regBody := map[string]any{
		"username":     "fdr_" + suf,
		"email":        "fdr_" + suf + "@meandproduct.dev",
		"password":     "Founder@2025ok!",
		"display_name": "Founder " + suf,
	}
	code, m, raw := postJSON("http://127.0.0.1:8080/api/v1/auth/register/self", "", regBody)
	if code == 409 {
		code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/auth/login", "", map[string]any{
			"username_or_email": "fdr_" + suf,
			"password":          "Founder@2025ok!",
		})
		if code != 200 {
			log.Fatalf("founder login failed code=%d body=%s", code, raw)
		}
	} else if code != 200 {
		log.Fatalf("founder register code=%d body=%s", code, raw)
	}
	d := asMap(m["data"])
	access := aStr(d["access_token"])
	refresh := aStr(d["refresh_token"])
	// DB 激活 pending
	d2 := db()
	user := asMap(d["user"])
	uid := uint64(aFloat(user["id"]))
	if uid > 0 {
		d2.Model(&entity.User{ID: uid}).UpdateColumn("status", entity.UserStatusActive)
	}
	return access, refresh
}

func mustJSON(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

func main() {
	suf := "t8"
	if len(os.Args) > 1 {
		suf = os.Args[1]
	}
	// 0. 等服务 up
	for i := 0; i < 30; i++ {
		code, _, _ := getURL("http://127.0.0.1:8080/api/v1/health", "")
		if code == 200 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	// 1. 准备 founder/admin 用户
	fdrAccess, _ := ensureFounderUser(suf)
	_, adminAccess, _ := ensureAdminUser(suf)

	// ============================================================
	// CP1: founder POST /founder/invites 生成 1 条邀请码（默认 max=1, 30d）
	// ============================================================
	code, m, raw := postJSON("http://127.0.0.1:8080/api/v1/founder/invites", fdrAccess, map[string]any{})
	check("CP1 founder 生成邀请码 → 200", code == 200 && aInt(m["code"]) == 0, "code="+strconv.Itoa(code), "body="+raw)
	var inviteCode string
	var inviteID float64
	if data := asMap(m["data"]); data != nil {
		inv := asMap(data["invite"])
		inviteCode = aStr(inv["code"])
		inviteID = aFloat(inv["id"])
	}
	check("CP1 返回 code.len=32 且 id>0", len(inviteCode) == 32 && inviteID > 0, "inviteCode="+inviteCode)

	// ============================================================
	// CP2: founder GET /founder/invites → items 至少 1 条且 created_by=用户自己
	// ============================================================
	code, m, raw = getURL("http://127.0.0.1:8080/api/v1/founder/invites?page=1&page_size=20", fdrAccess)
	check("CP2 founder 列表 200 code=0", code == 200 && aInt(m["code"]) == 0, "body="+raw)
	items := asSlice(asMap(m["data"])["items"])
	total := aFloat(asMap(m["data"])["total"])
	check("CP2 total>=1 && items>=1", total >= 1 && len(items) >= 1)

	// ============================================================
	// CP3: founder 作废自己的 invite → 200 code=0 ；/auth/invite/validate 变为 revoked usable=false
	// ============================================================
	revBody := map[string]any{"invite_id": uint64(inviteID)}
	code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/founder/invites/revoke", fdrAccess, revBody)
	check("CP3 founder revoke → 200", code == 200 && aInt(m["code"]) == 0, "body="+raw)
	code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/auth/invite/validate", "", map[string]any{"invite_code": inviteCode})
	d := asMap(m["data"])
	usable := true
	if u, ok := d["usable"].(bool); ok {
		usable = u
	}
	msg := aStr(d["message"])
	check("CP3 ValidateInvite → revoked usable=false msg contains revoked",
		code == 200 && !usable && strings.Contains(strings.ToLower(msg), "revoke"),
		"body="+raw)

	// ============================================================
	// CP4: admin 批量生成 count=5 max_uses=3 → items 5 个 各 max=3 status=active created_by=adminUid
	// ============================================================
	code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/admin/invites/generate", adminAccess, map[string]any{
		"count":        5,
		"max_uses":     3,
		"expires_days": 60,
	})
	check("CP4 admin generate count=5 → 200 len(items)=5", code == 200 && len(asSlice(asMap(m["data"])["items"])) == 5, "body="+raw)
	adminItems := asSlice(asMap(m["data"])["items"])
	ids := []uint64{}
	for _, it := range adminItems {
		inv := asMap(it)
		check("CP4 每个 max=3 active", aFloat(inv["max_uses"]) == 3 && aStr(inv["status"]) == "active")
		ids = append(ids, uint64(aFloat(inv["id"])))
	}

	// ============================================================
	// CP5: admin GET /admin/invites?status=active → items 至少 5+1=6 ？包含 CP1（revoked 不计） + CP4 5 active
	// ============================================================
	code, m, raw = getURL("http://127.0.0.1:8080/api/v1/admin/invites?status=active&page=1&page_size=20", adminAccess)
	check("CP5 admin active list 200 total>=5", code == 200 && aFloat(asMap(asMap(m["data"]))["total"]) >= 5, "body="+raw)

	// ============================================================
	// CP6: admin revoke ids 2个 → revoked=2；validate 查第一条 status=revoked
	// ============================================================
	if len(ids) >= 2 {
		code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/admin/invites/revoke", adminAccess, map[string]any{"ids": []uint64{ids[0], ids[1]}})
		check("CP6 admin revoke many rows=2", code == 200 && aFloat(asMap(m["data"])["revoked"]) == 2, "body="+raw)
		var firstCode string
		for _, it := range adminItems[:1] {
			firstCode = aStr(asMap(it)["code"])
		}
		if firstCode != "" {
			code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/auth/invite/validate", "", map[string]any{"invite_code": firstCode})
			d := asMap(m["data"])
			usable := true
			if u, ok := d["usable"].(bool); ok {
				usable = u
			}
			msg := aStr(d["message"])
			check("CP6 validate firstCode usable=false msg contains revoked",
				code == 200 && !usable && strings.Contains(strings.ToLower(msg), "revoke"),
				"body="+raw)
		}
	}

	// ============================================================
	// RBAC 验证
	// ============================================================
	// CP7: founder 访问 /admin/invites → 403
	code, _, raw = getURL("http://127.0.0.1:8080/api/v1/admin/invites", fdrAccess)
	check("CP7 founder 访问 admin → 403", code == 403, "code="+strconv.Itoa(code)+" body="+raw)

	// CP8: guest（无 token）访问 /founder/invites → 401
	code, _, raw = getURL("http://127.0.0.1:8080/api/v1/founder/invites", "")
	check("CP8 guest → 401", code == 401, "code="+strconv.Itoa(code)+" body="+raw)

	// CP9: admin 访问 founder 路由（RBAC(founder,admin)） → 200
	code, _, raw = getURL("http://127.0.0.1:8080/api/v1/founder/invites?page=1&page_size=5", adminAccess)
	check("CP9 admin 代访问 founder/invites → 200", code == 200, "code="+strconv.Itoa(code)+" body="+raw)

	// CP10: DTO 非法 generate max_uses=0 expires_days=0 → 422 detail>=2
	code, m, raw = postJSON("http://127.0.0.1:8080/api/v1/founder/invites", fdrAccess, map[string]any{
		"max_uses":     0,
		"expires_days": 0,
	})
	_ = context.Background()
	check("CP10 DTO非法 founder generate → 422", code == 422, "code="+strconv.Itoa(code)+" body="+raw)

	fmt.Printf("\n==== RESULT: PASS %d / FAIL %d ====\n", pass, fail)
	os.Exit(fail)
}
