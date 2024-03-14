package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

var (
	adminUsername     = "admin"
	adminUserPassword = "1qazXSW@"
)

// 为http.HandlerFunc封装basic auth功能，并返回
func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求头中出去用户和密码，如果认证请求头不存在或值无效，则ok为false
		username, password, ok := r.BasicAuth()
		if ok {
			// SHA-256算法计算用户名和密码
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(adminUsername))
			expectedPasswordHash := sha256.Sum256([]byte(adminUserPassword))

			// subtle.ConstantTimeCompare() 函数检查用户名和密码，相等返回1，否则返回0；
			// 重要的是先进行hash使得两者长度相等，从而避免泄露信息。
			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			// 如果比较结果正确，调用下一个处理器。
			// 最后调用return，为了让后面代码不被执行
			if usernameMatch && passwordMatch {
				// next.ServeHTTP(w,r) = next(w,r)
				next.ServeHTTP(w, r)
				return
			}
		}
		// 如果认证头不存在或无效，亦或凭证不正确，
		// 则需要设置WWW-Authenticate头信息并发送401未认证响应，为了通知客户端使用基本认证
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
