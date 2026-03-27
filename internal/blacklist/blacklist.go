package blacklist

import (
	"net"
	"strings"
	"sync"
	"time"
)

// Manager 黑名单管理器
type Manager struct {
	mu            sync.RWMutex
	userBlacklist map[string]*banInfo    // 用户名黑名单
	ipBlacklist   map[string]*banInfo    // IP黑名单
	ipUserMap     map[string]map[string]time.Time // IP->用户名->首次请求时间
	ipBanThreshold int                    // IP封禁阈值（同IP换用户名数量）
	ipCheckWindow  time.Duration          // IP检测时间窗口
	banDuration    time.Duration          // 封禁时长
}

// banInfo 封禁信息
type banInfo struct {
	bannedAt   time.Time
	reason     string
	expireAt   time.Time
}

// New 创建黑名单管理器
func New() *Manager {
	return &Manager{
		userBlacklist:  make(map[string]*banInfo),
		ipBlacklist:    make(map[string]*banInfo),
		ipUserMap:      make(map[string]map[string]time.Time),
		ipBanThreshold: 5,                     // 同IP 5个不同用户名触发封禁
		ipCheckWindow:  10 * time.Minute,       // 10分钟检测窗口
		banDuration:    24 * time.Hour,         // 默认封禁24小时
	}
}

// NewWithOptions 使用自定义参数创建黑名单管理器
func NewWithOptions(ipBanThreshold int, ipCheckWindow, banDuration time.Duration) *Manager {
	return &Manager{
		userBlacklist:  make(map[string]*banInfo),
		ipBlacklist:    make(map[string]*banInfo),
		ipUserMap:      make(map[string]map[string]time.Time),
		ipBanThreshold: ipBanThreshold,
		ipCheckWindow:  ipCheckWindow,
		banDuration:    banDuration,
	}
}

// IsUserBanned 检查用户名是否在黑名单中
func (m *Manager) IsUserBanned(username string) (bool, string, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	username = strings.ToLower(strings.TrimSpace(username))
	info, exists := m.userBlacklist[username]
	if !exists {
		return false, "", time.Time{}
	}

	// 检查是否过期
	if time.Now().After(info.expireAt) {
		return false, "", time.Time{}
	}

	return true, info.reason, info.expireAt
}

// IsIPBanned 检查IP是否在黑名单中
func (m *Manager) IsIPBanned(ip string) (bool, string, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ip = m.normalizeIP(ip)
	info, exists := m.ipBlacklist[ip]
	if !exists {
		return false, "", time.Time{}
	}

	// 检查是否过期
	if time.Now().After(info.expireAt) {
		return false, "", time.Time{}
	}

	return true, info.reason, info.expireAt
}

// BanUser 将用户名加入黑名单
func (m *Manager) BanUser(username, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	username = strings.ToLower(strings.TrimSpace(username))
	now := time.Now()
	m.userBlacklist[username] = &banInfo{
		bannedAt: now,
		reason:   reason,
		expireAt: now.Add(m.banDuration),
	}
}

// BanIP 将IP加入黑名单
func (m *Manager) BanIP(ip, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ip = m.normalizeIP(ip)
	now := time.Now()
	m.ipBlacklist[ip] = &banInfo{
		bannedAt: now,
		reason:   reason,
		expireAt: now.Add(m.banDuration),
	}
}

// RecordIPUser 记录IP访问的用户名，返回是否应该封禁该IP
func (m *Manager) RecordIPUser(ip, username string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	ip = m.normalizeIP(ip)
	username = strings.ToLower(strings.TrimSpace(username))
	now := time.Now()

	// 清理过期记录
	m.cleanupExpiredRecords(ip, now)

	// 初始化IP记录
	if m.ipUserMap[ip] == nil {
		m.ipUserMap[ip] = make(map[string]time.Time)
	}

	// 记录新用户名（如果已存在则不更新时间，保留首次请求时间）
	if _, exists := m.ipUserMap[ip][username]; !exists {
		m.ipUserMap[ip][username] = now
	}

	// 检查是否超过阈值
	shouldBan := len(m.ipUserMap[ip]) >= m.ipBanThreshold
	return shouldBan
}

// GetIPUserCount 获取某个IP访问的不同用户名数量
func (m *Manager) GetIPUserCount(ip string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ip = m.normalizeIP(ip)
	return len(m.ipUserMap[ip])
}

// cleanupExpiredRecords 清理过期的IP-用户名记录
func (m *Manager) cleanupExpiredRecords(ip string, now time.Time) {
	if userMap, exists := m.ipUserMap[ip]; exists {
		for user, firstSeen := range userMap {
			if now.Sub(firstSeen) > m.ipCheckWindow {
				delete(userMap, user)
			}
		}
		if len(userMap) == 0 {
			delete(m.ipUserMap, ip)
		}
	}
}

// normalizeIP 规范化IP地址
func (m *Manager) normalizeIP(ip string) string {
	// 移除端口号
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	// 统一转为小写（IPv6）
	return strings.ToLower(strings.TrimSpace(ip))
}

// CleanExpired 清理所有过期的封禁记录和IP映射
func (m *Manager) CleanExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 清理用户黑名单
	for user, info := range m.userBlacklist {
		if now.After(info.expireAt) {
			delete(m.userBlacklist, user)
		}
	}

	// 清理IP黑名单
	for ip, info := range m.ipBlacklist {
		if now.After(info.expireAt) {
			delete(m.ipBlacklist, ip)
		}
	}

	// 清理IP-用户名映射
	for ip, userMap := range m.ipUserMap {
		for user, firstSeen := range userMap {
			if now.Sub(firstSeen) > m.ipCheckWindow {
				delete(userMap, user)
			}
		}
		if len(userMap) == 0 {
			delete(m.ipUserMap, ip)
		}
	}
}

// GetStats 获取黑名单统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"userBlacklistCount": len(m.userBlacklist),
		"ipBlacklistCount":   len(m.ipBlacklist),
		"ipTrackingCount":    len(m.ipUserMap),
		"banDuration":        m.banDuration.String(),
		"ipCheckWindow":      m.ipCheckWindow.String(),
		"ipBanThreshold":     m.ipBanThreshold,
	}
}

// UnbanUser 手动解封用户（管理员功能）
func (m *Manager) UnbanUser(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	username = strings.ToLower(strings.TrimSpace(username))
	delete(m.userBlacklist, username)
}

// UnbanIP 手动解封IP（管理员功能）
func (m *Manager) UnbanIP(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ip = m.normalizeIP(ip)
	delete(m.ipBlacklist, ip)
}
