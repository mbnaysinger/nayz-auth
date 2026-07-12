package middlewares

import "net/http"

// CorsMiddleware gerencia o compartilhamento de recursos entre diferentes portas/domínios.
// o navegador, por segurança, envia uma requisição prévia fantasma do tipo "OPTIONS" (Preflight).
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Liberamos as origens, os métodos e os cabeçalhos (como o Authorization que carrega o JWT)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Em produção, mude "*" para o domínio do Frontend
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
