package handler

import (
	"Audiophile/database"
	"Audiophile/database/helper"
	"Audiophile/models"
	"Audiophile/utilities"
	"context"
	"database/sql"
	firebase "firebase.google.com/go"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var JwtKey = []byte("secret_key")

func Login(w http.ResponseWriter, r *http.Request) {
	var userDetails models.UsersLoginDetails
	decoderErr := utilities.Decoder(r, &userDetails)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	if userDetails.Email == "" {
		OauthLogin(userDetails.OauthToken, w, r)
		return
	}

	userDetails.Email = strings.ToLower(userDetails.Email)

	userCredentials, fetchErr := helper.FetchPasswordAndIDANDRole(userDetails.Email, userDetails.Role)

	if fetchErr != nil {
		if fetchErr == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("ERROR: Wrong details"))
			if err != nil {
				return
			}

			logrus.Printf("FetchPasswordAndId: not able to get password, id, or role:%v", fetchErr)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if PasswordErr := bcrypt.CompareHashAndPassword([]byte(userCredentials.Password), []byte(userDetails.Password)); PasswordErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logrus.Printf("password misMatch")
		_, err := w.Write([]byte("ERROR: Wrong password"))
		if err != nil {
			return
		}
		return
	}

	expiresAt := time.Now().Add(60 * time.Minute)

	claims := &models.Claims{
		ID:   userCredentials.ID,
		Role: userCredentials.Role,
		StandardClaims: jwt.StandardClaims{

			ExpiresAt: expiresAt.Unix(),
			// Issuer:    userCredentials.Role,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("TokenString: cannot create token string:%v", err)
		return
	}

	err = helper.CreateSession(claims)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("CreateSession: cannot create session:%v", err)
		return
	}

	userOutboundData := make(map[string]interface{})

	userOutboundData["token"] = tokenString

	err = utilities.Encoder(w, userOutboundData)
	if err != nil {
		logrus.Printf("Login: Not able to login:%v", err)
		return
	}
}

func OauthLogin(oauthToken string, w http.ResponseWriter, r *http.Request) {
	opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Printf("OAuth:cannot create firebase application object:%v", err)
		return
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		logrus.Printf("OauthLogin:cannot create client client:%v", err)
	}

	//header := r.Header.Get(echo.HeaderAuthorization)
	idToken := strings.TrimSpace(strings.Replace(oauthToken, "Bearer", "", 1))
	firebaseToken, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("OauthLogin:cannot verify token:%v", err)
		return
	}
	userDetails, err := client.GetUser(context.Background(), firebaseToken.UID)
	if err != nil {
		logrus.Printf("OauthLogin: cannot get user details:%v", err)
		return
	}

	err = helper.CheckEmail(userDetails.Email)
	var userID uuid.UUID
	if err != nil {
		if err == sql.ErrNoRows {
			ID, createErr := helper.CreateNewUser(userDetails)
			if createErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Printf("OauthLogin: cannot create new user:%v", createErr)
				return
			}
			userID = ID
		} else {
			logrus.Printf("OauthLogin: cannot check user google email:%v", err)
			return
		}
	}

	expiresAt := time.Now().Add(60 * time.Minute)

	claims := &models.Claims{
		ID:   userID,
		Role: "user",
		StandardClaims: jwt.StandardClaims{

			ExpiresAt: expiresAt.Unix(),
			// Issuer:    userCredentials.Role,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("TokenString: cannot create token string:%v", err)
		return
	}

	err = helper.CreateSession(claims)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("CreateSession: cannot create session:%v", err)
		return
	}

	userOutboundData := make(map[string]interface{})

	userOutboundData["token"] = tokenString

	err = utilities.Encoder(w, userOutboundData)
	if err != nil {
		logrus.Printf("Login: Not able to login:%v", err)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress:QueryParam for ID:%v", ok)
		return
	}

	err := helper.Logout(contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("Logout:unable to logout:%v", err)
		return
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	var userDetails models.Users

	decoderErr := utilities.Decoder(r, &userDetails)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userID, err := helper.CreateUser(&userDetails, tx)
		if err != nil {
			logrus.Printf("Register:CreateUser:%v", err)
			return err
		}
		userDetails.ID = userID
		err = helper.CreateAddress(userID, &userDetails, tx)
		if err != nil {
			logrus.Printf("Register:CreateAddress:%v", err)
			return err
		}
		err = helper.CreateRole(userID, models.UserRoleUser, tx)
		return err
	})
	if txErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("Register:%v", txErr)
		return
	}

	userOutboundData := make(map[string]uuid.UUID)

	userOutboundData["Successfully Registered: ID is"] = userDetails.ID

	err := utilities.Encoder(w, userOutboundData)
	if err != nil {
		logrus.Printf("Register:%v", err)
		return
	}
}

func filters(r *http.Request) (models.FiltersCheck, error) {
	filtersCheck := models.FiltersCheck{}
	isSearched := false
	searchedName := r.URL.Query().Get("name")
	if searchedName != "" {
		isSearched = true
	}

	var limit int
	var err error
	var page int
	strLimit := r.URL.Query().Get("limit")
	if strLimit == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(strLimit)
		if err != nil {
			logrus.Printf("Limit: cannot get limit:%v", err)
			return filtersCheck, err
		}
	}

	strPage := r.URL.Query().Get("page")
	if strPage == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(strPage)
		if err != nil {
			logrus.Printf("Page: cannot get page:%v", err)
			return filtersCheck, err
		}
	}

	filtersCheck = models.FiltersCheck{
		IsSearched:   isSearched,
		SearchedName: searchedName,
		Page:         page,
		Limit:        limit}
	return filtersCheck, nil
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	filterCheck, err := filters(r)
	if err != nil {
		logrus.Printf("GetUsers:filterCheck:%v", err)
	}

	adminGetUserDetails, AdminGetUserErr := helper.GetUsers(filterCheck)
	if AdminGetUserErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("GetUser: not able to get users :%v ", AdminGetUserErr)
		return
	}

	err = utilities.Encoder(w, adminGetUserDetails)
	if err != nil {
		logrus.Printf("GetUser:%v", err)
		return
	}
}

func AddCategory(w http.ResponseWriter, r *http.Request) {
	var categoryDetails models.Categories

	decoderErr := utilities.Decoder(r, &categoryDetails)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	categoryID, err := helper.AddCategory(&categoryDetails)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddCategory:cannot add category:%v", err)
		return
	}

	userOutboundData := make(map[string]uuid.UUID)

	userOutboundData["Successfully Added Category: ID is"] = categoryID

	err = utilities.Encoder(w, userOutboundData)
	if err != nil {
		logrus.Printf("AddCategory:%v", err)
		return
	}
}

func AddBrands(w http.ResponseWriter, r *http.Request) {
	var brandDetails models.Brands

	decoderErr := utilities.Decoder(r, &brandDetails)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	err := helper.AddBrands(&brandDetails)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddBrands: cannot add brand:%v", err)
		return
	}

	message := "Successfully added brand"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("AddProduct:%v", err)
		return
	}
}

func AddProduct(w http.ResponseWriter, r *http.Request) {
	productDetails := make([]models.Product, 0)

	decoderErr := utilities.Decoder(r, &productDetails)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	if len(productDetails) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("AddProduct: products cannot be empty")
		return
	}

	categoryID := r.URL.Query().Get("categoryID")

	err := helper.AddProduct(productDetails, categoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddProduct:cannot add product to inventory:%v", err)
		return
	}

	message := "Successfully added product to inventory"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("AddProduct:%v", err)
		return
	}
}

func AddProductImages(w http.ResponseWriter, r *http.Request) {
	productImages := make([]models.ProductImages, 0)

	err := utilities.Decoder(r, &productImages)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder Error:%v", err)
		return
	}

	if len(productImages) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("BulkInsertDishes: Dishes cannot be empty")
		return
	}

	productID := chi.URLParam(r, "productID")

	err = helper.AddProductImages(productImages, productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddProductImages: not able to add images for a product:%v", err)
		return
	}
	message := "Successfully added images for the product"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("encoder error:%v", err)
		return
	}
}

func ViewProducts(w http.ResponseWriter, r *http.Request) {
	filterCheck, err := filters(r)
	if err != nil {
		logrus.Printf("ViewProduct: filterCheck error:%v", err)
		return
	}

	productDetails, productDetailsErr := helper.ViewProducts(filterCheck)
	if productDetailsErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("ViewProducts: not able to get productDetails: %v", productDetailsErr)
		return
	}

	err = utilities.Encoder(w, productDetails)
	if err != nil {
		logrus.Printf("ViewProduct: %v", err)
		return
	}
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productID")

	var productDetails models.ProductUpdateDetails
	decoderErr := utilities.Decoder(r, &productDetails)
	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	updateProductErr := helper.UpdateProduct(productID, productDetails)
	if updateProductErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("UpdateProduct: not able to update product:%v", updateProductErr)
		return
	}

	message := "updated product successfully"
	err := utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("UpdateProduct:%v", err)
		return
	}
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productID")

	err := helper.DeleteProduct(productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("DeleteProduct: Unable to delete product:%v", err)
		return
	}

	message := "deleted product successfully"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("UpdateProduct:%v", err)
		return
	}
}

func DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	productImageID := chi.URLParam(r, "productImageID")

	err := helper.DeleteProductImage(productImageID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("DeleteProductImage: Unable to delete product:%v", err)
		return
	}

	message := "deleted product image successfully"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("DeleteProductImage:%v", err)
		return
	}
}

func AddAddress(w http.ResponseWriter, r *http.Request) {
	var userDetails models.Users

	decoderErr := utilities.Decoder(r, &userDetails)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress:QueryParam for ID:%v", ok)
		return
	}

	err := helper.AddAddress(contextValues.ID, &userDetails)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress: cannot add address:%v", err)
		return
	}

	message := "Added address successfully"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("AddAddress:%v", err)
		return
	}
}

func AddToCart(w http.ResponseWriter, r *http.Request) {
	var quantity models.Quantity

	decoderErr := utilities.Decoder(r, &quantity)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress:QueryParam for ID:%v", ok)
		return
	}

	productID := chi.URLParam(r, "productID")

	price, err := helper.FetchPrice(productID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("AddToCart:unable to get price of product:%v", err)
		return
	}

	err = helper.AddToCart(productID, quantity, price, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("AddToCart: cannot add product to cart:%v", err)
		return
	}

	message := "Successfully added product to cart"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("AddToCart:%v", err)
		return
	}
}

func RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	cartID := chi.URLParam(r, "cartID")

	err := helper.RemoveFromCart(cartID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("RemoveFromCart: Unable to remove product:%v", err)
		return
	}

	message := "removed product from cart successfully"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("RemoveFromCart:%v", err)
		return
	}
}

func SelectProduct(w http.ResponseWriter, r *http.Request) {
	cartID := r.URL.Query().Get("cartId")

	err := helper.SelectProduct(cartID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("SelectProducts:cannot select product:%v", err)
		return
	}

	message := "Selected product successfully"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("RemoveFromCart:%v", err)
		return
	}
}

func CheckOut(w http.ResponseWriter, r *http.Request) {
	var orderDetails models.OrderDetails

	//decoderErr := utilities.Decoder(r, &orderDetails)
	//
	//if decoderErr != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	logrus.Printf("Decoder error:%v", decoderErr)
	//	return
	//}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress:QueryParam for ID:%v", ok)
		return
	}

	addressID, err := helper.FetchUserAddress(contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("CheckOut:FetchUserAddress:%v", err)
		return
	}

	orderedProducts, err := helper.FetchOrderedProducts(contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("FetchOrderedProducts: cannot get cart id's:%v", err)
		return
	}
	orderDetails.CartID = orderedProducts

	totalAmount := 0.0
	Amount := 0.0
	//cartID := chi.URLParam(r, "cartID")

	for _, post := range orderDetails.CartID {
		Amount, err = helper.FetchTotalAmount(post, contextValues.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("FetchTotalAmount: not able to fetch toatl amount:%v", err)
			return
		}
		totalAmount = totalAmount + Amount
	}

	orderDetails.UserID = contextValues.ID
	orderDetails.AddressID = addressID
	orderDetails.TotalAmount = totalAmount

	//err = utilities.Decoder(r, &orderDetails)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	logrus.Printf("Decoder Error:%v", err)
	//	return
	//}

	err = helper.CheckOut(orderDetails)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("CheckOut: error is:%v", err)
		return
	}

	message := "CheckOut Successful"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("CheckOut:%v", err)
		return
	}
}

func InstantPayment(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress:QueryParam for ID:%v", ok)
		return
	}

	var paymentDetails models.PaymentDetails
	err := utilities.Decoder(r, &paymentDetails)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder Error:%v", err)
		return
	}
	// transaction begin
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		paymentID, err := helper.InstantPayment(contextValues.ID, paymentDetails, orderID, tx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("InstantPayment: cannot make payment:%v", err)
			return err
		}

		err = helper.UpdateOrder(orderID, tx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("InstantPayment:UpdateOrder: cannot complete payment:%v", err)
			return err
		}

		err = helper.CreateBill(contextValues.ID, paymentID, orderID, tx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("InstantPayment:CreateBill:%v", err)
			return err
		}

		err = helper.RemoveProductsFromCart(contextValues.ID, tx)
		return err
	})
	if txErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("InstantPayment:%v", txErr)
		return
	}
	// transaction ended

	message := "Completed Payment Successfully"
	err = utilities.Encoder(w, message)
	if err != nil {
		logrus.Printf("InstantPayment:%v", err)
		return
	}
}

func ViewBillDetails(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("AddAddress:QueryParam for ID:%v", ok)
		return
	}
	billDetails, err := helper.ViewBillDetails(contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("ViewBillDetails: cannot get bill details:%v", err)
		return
	}

	err = utilities.Encoder(w, &billDetails)
	if err != nil {
		logrus.Printf("InstantPayment:%v", err)
		return
	}
}
