package main

import (
	"log"
	"net/http"

	"github.com/souzavini/GoCrudRestApi/configs"
	"github.com/souzavini/GoCrudRestApi/internal/entity"
	"github.com/souzavini/GoCrudRestApi/internal/infra/database"
	"github.com/souzavini/GoCrudRestApi/internal/infra/webserver/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	_ "github.com/souzavini/GoCrudRestApi/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           Go Expert API Example
// @version         1.0
// @description     Product API with auhtentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   Vinicius Souza
// @contact.url    http://www.vinicius.souza.com.br
// @contact.email  atendimento@vinicius.com.br

// @license.name   Vini License
// @license.url    http://www.vinicius.souza.com.br

// @host      localhost:8000
// @BasePath  /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&entity.Product{}, &entity.User{})

	productDB := database.NewProduct(db)
	productHandler := handlers.NewProductHandler(productDB)

	userDb := database.NewUser(db)
	userHandler := handlers.NewUserHandler(userDb)

	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)                           //Absorver Panics e apicacao nao cai
	r.Use(middleware.WithValue("jwt", configs.TokenAuth)) //Injeta instancia do token no context
	r.Use(middleware.WithValue("JwtExperesIn", configs.JwtExperesIn))

	r.Use(LogRequest) //Midlewarre proprio para logar as requisições

	r.Route("/products", func(r chi.Router) {
		r.Use((jwtauth.Verifier(configs.TokenAuth))) //Injeta instancia do token no context
		r.Use(jwtauth.Authenticator)                 //Valida se o token e valido com base no nosso SECRET

		r.Post("/", productHandler.CreateProduct)
		r.Put("/{id}", productHandler.UpdateProduct)
		r.Delete("/{id}", productHandler.DeleteProduct)
		r.Get("/", productHandler.GetProducts)
		r.Get("/{id}", productHandler.GetProduct)
	})

	r.Post("/users", userHandler.Create)
	r.Post("/users/generate_token", userHandler.GetJWT)

	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8000/docs/doc.json")))

	http.ListenAndServe(":8000", r)
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
