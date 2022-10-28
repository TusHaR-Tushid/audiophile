package server

import (
	"Audiophile/handler"
	"Audiophile/middleware"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server struct {
	chi.Router
}

func SetupRoutes() *Server {
	router := chi.NewRouter()
	router.Route("/audiophile", func(audiophile chi.Router) {
		audiophile.Route("/health", func(r chi.Router) {
			r.Get("/api", func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprintf(w, "Remote State Team")
				if err != nil {
					return
				}
			})
		})
		audiophile.Get("/", handler.ViewProducts)
		audiophile.Post("/register", handler.Register)
		audiophile.Post("/log-in", handler.Login)
		audiophile.Put("/log-out", handler.Logout)
		audiophile.Route("/auth", func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware)
			auth.Post("/address", handler.AddAddress)
			auth.Post("/{productID}/cart", handler.AddToCart)
			auth.Delete("/{cartID}/cart", handler.RemoveFromCart)
			auth.Post("/image", handler.UploadImage)
			auth.Post("/", handler.SelectProduct)
			auth.Post("/checkout", handler.CheckOut)
			auth.Post("/{orderID}/payment", handler.InstantPayment)
			auth.Post("/bill", handler.ViewBillDetails)
			auth.Put("/log-out", handler.Logout)
			auth.Route("/admin", func(admin chi.Router) {
				admin.Use(middleware.AdminMiddleware)
				admin.Get("/users", handler.GetUsers)
				admin.Post("/category", handler.AddCategory)
				admin.Post("/brand", handler.AddBrands)
				admin.Post("/inventory", handler.AddProduct)
				admin.Get("/products", handler.ViewProducts)
				admin.Route("/{productID}", func(product chi.Router) {
					product.Post("/product-images", handler.AddProductImages)
					product.Put("/", handler.UpdateProduct)
					product.Delete("/", handler.DeleteProduct)
				})
				admin.Delete("/{productImageID}", handler.DeleteProductImage)
			})
		})
	})
	return &Server{router}
}

func (svc *Server) Run(port string) error {
	return http.ListenAndServe(port, svc)
}
