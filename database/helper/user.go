package helper

import (
	"Audiophile/database"
	"Audiophile/models"
	"github.com/elgris/sqrl"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func FetchPasswordAndIDANDRole(userMail, userRole string) (models.UserCredentials, error) {
	SQL := `SELECT users.id,password,role
            FROM   users
            JOIN   roles ON users.id=roles.user_id
            WHERE  email=$1
            AND    roles.role=$2 
            AND    archived_at IS NULL `

	var userCredentials models.UserCredentials

	err := database.AudiophileDB.Get(&userCredentials, SQL, userMail, userRole)
	if err != nil {
		logrus.Printf("FetchPassword: Not able to fetch password, ID or role: %v", err)
		return userCredentials, err
	}
	return userCredentials, nil
}

func CreateSession(claims *models.Claims) error {
	SQL := `INSERT INTO sessions(user_id)
            VALUES   ($1)`
	_, err := database.AudiophileDB.Exec(SQL, claims.ID)
	if err != nil {
		logrus.Printf("CreateSession: cannot create user session:%v", err)
		return err
	}
	return nil
}

func Logout(userID uuid.UUID) error {
	SQL := `UPDATE sessions
            SET    expires_at=now()
            WHERE  user_id=$1`

	_, err := database.AudiophileDB.Exec(SQL, userID)
	if err != nil {
		logrus.Printf("Logout: cannot do logout:%v", err)
		return err
	}
	return nil
}

func CheckSession(userID uuid.UUID) (uuid.UUID, error) {
	SQL := `SELECT id
           FROM    sessions
           WHERE   expires_at IS NULL
           AND     user_id=$1`
	var sessionID uuid.UUID

	err := database.AudiophileDB.Get(&sessionID, SQL, userID)
	if err != nil {
		logrus.Printf("CheckSession: session expired:%v", err)
		return sessionID, err
	}
	return sessionID, nil
}

func CreateUser(userDetails *models.Users, tx *sqlx.Tx) (uuid.UUID, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userDetails.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Printf("CreateUser: Not able to hash password:%v", err)
		return userDetails.ID, err
	}

	// language=SQL
	SQL := `INSERT  INTO  users(name, email, password, phone_no, age)
           VALUES  ($1, $2, $3, $4, $5)
           RETURNING id`

	var id uuid.UUID
	userDetails.Email = strings.ToLower(userDetails.Email)

	err = tx.Get(&id, SQL, userDetails.Name, userDetails.Email, hashPassword, userDetails.PhoneNo, userDetails.Age)

	if err != nil {
		logrus.Printf("CreateUser: Not able to Create User :%v", err)
		return id, err
	}

	return id, nil
}

func CreateAddress(id uuid.UUID, userDetails *models.Users, tx *sqlx.Tx) error {
	SQL := `INSERT INTO user_address(user_id, address)
          VALUES ($1, $2)`
	_, err := tx.Exec(SQL, id, userDetails.Address)

	if err != nil {
		logrus.Printf("CreateAddress: unable to create address:%v", err)
		return err
	}

	return nil
}

func CreateRole(id uuid.UUID, role models.UserProfile, tx *sqlx.Tx) error {
	SQL := `INSERT INTO roles(user_id, role)
           VALUES ($1, $2)`

	_, err := tx.Exec(SQL, id, role)

	if err != nil {
		logrus.Printf("CreateRoleUser: Not able to set user role: %v", err)
		return err
	}

	return nil
}

func UploadImage(imageURL string) (string, error) {
	// language=SQL
	var imageID string
	SQL := `INSERT INTO images(url)
          VALUES  ($1)
          RETURNING id`
	err := database.AudiophileDB.Get(&imageID, SQL, imageURL)
	if err != nil {
		logrus.Printf("UploadImage: not able to store image in db:%v", err)
		return imageID, err
	}
	return imageID, nil
}

func GetUsers(filterCheck models.FiltersCheck) (models.TotalUser, error) {
	var totalUser models.TotalUser
	// language=SQL
	SQL := `WITH  cte_User AS(

            SELECT  
                     count(*) over () as total_count,
                     users.id as id,
                     name,
                     email,
                     phone_no,
                     password,
                     age,
                     users.created_at as created_at,
                     users.updated_at as updated_at,
                     role,
                     address
            FROM  users 
            JOIN roles ON users.id=roles.user_id
            JOIN   user_address ON roles.user_id=user_address.user_id
            WHERE users.archived_at IS NULL
            AND   user_address.archived_at IS NULL
            AND   role=$1
            AND    ($2 or name ilike '%'||$3||'%')
            ORDER BY name
            LIMIT $4 OFFSET $5
            )
            SELECT total_count,
                      id,
                     name,
                     email,
                     phone_no,
                     password,
                     age,
                     created_at,
                     updated_at,
                     role,
                     address
            FROM     cte_User`

	userDetails := make([]models.UserDetails, 0)

	err := database.AudiophileDB.Select(&userDetails, SQL, models.UserRoleUser, !filterCheck.IsSearched, filterCheck.SearchedName, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)

	if err != nil {
		logrus.Printf("GetUsers: unable to fetch user details:%v", err)
		return totalUser, err
	}

	if len(userDetails) == 0 {
		logrus.Printf("GetUsers:%v", err)
		return totalUser, err
	}

	totalUser.UserDetails = userDetails
	totalUser.TotalCount = userDetails[0].TotalCount
	return totalUser, nil
}

func AddBrands(brandDetails *models.Brands) error {
	SQL := `INSERT INTO brands(brand_name, brand_description)
             VALUES   ($1, $2)`

	_, err := database.AudiophileDB.Exec(SQL, brandDetails.BrandName, brandDetails.BrandDescription)
	if err != nil {
		logrus.Printf("AddBrands: cannot add brands:%v", err)
		return err
	}
	return nil
}

func AddCategory(categoryDetails *models.Categories) (uuid.UUID, error) {
	SQL := `INSERT INTO category(name)
            VALUES ($1)
            RETURNING id`
	var categoryID uuid.UUID
	err := database.AudiophileDB.Get(&categoryID, SQL, categoryDetails.Name)
	if err != nil {
		logrus.Printf("AddCategory: cannot add category: %v", err)
		return categoryID, err
	}
	return categoryID, nil
}

func AddProduct(productDetails []models.Product, categoryID string) error {
	psql := sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)
	sql := psql.Insert("inventory").Columns("name", "price", "quantity", "category_id", "brand_id", "product_description")
	for _, post := range productDetails {
		sql.Values(post.Name, post.Price, post.Quantity, categoryID, post.BrandID, post.ProductDescription)
	}

	SQL, args, err := sql.ToSql()
	if err != nil {
		logrus.Printf("AddProduct: not able to create sql string: %v", err)
		return err
	}

	_, err = database.AudiophileDB.Exec(SQL, args...)
	if err != nil {
		logrus.Printf("AddProduct: not able to add product to inventory:%v", err)
		return err
	}

	return nil
}

func AddProductImages(productImagesDetails []models.ProductImages, productID string) error {
	psql := sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)
	sql := psql.Insert("images_per_product").Columns("image_id", "product_id")
	for _, post := range productImagesDetails {
		sql.Values(post.ImageID, productID)
	}

	SQL, args, err := sql.ToSql()
	if err != nil {
		logrus.Printf("AddProductImages: not able to create sql string: %v", err)
		return err
	}

	_, err = database.AudiophileDB.Exec(SQL, args...)
	if err != nil {
		logrus.Printf("AddProductImages: not able to add product images:%v", err)
		return err
	}

	return nil
}

func ViewProducts(filterCheck models.FiltersCheck) (models.TotalProduct, error) {
	var totalProducts models.TotalProduct

	SQL := `WITH cte_inventory AS(
            
            SELECT 
                     count(*) over () as total_count,
                     inventory.id as id,
                     name,
                     category_id,
                     price,
                     quantity,
                     url,
                     brand_id
            FROM   inventory
                        LEFT JOIN images_per_product ON inventory.id = product_id
                        LEFT JOIN images ON image_id = images.id
            WHERE inventory.archived_at IS NULL
            AND    ($1 or name ilike '%' || $2 || '%')
            AND    inventory.archived_at IS NULL 
            ORDER BY name
            LIMIT  $3 OFFSET $4
            
            )
           
            SELECT   total_count,
                     id,
                     name,
                     category_id,
                     price,
                     quantity,
                     COALESCE(url, '') as url
            FROM 
                     cte_inventory`

	productDetails := make([]models.ProductDetails, 0)

	err := database.AudiophileDB.Select(&productDetails, SQL, !filterCheck.IsSearched, filterCheck.SearchedName, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)
	if err != nil {
		logrus.Printf("ViewProducts: unable to fetch product details:%v", err)
		return totalProducts, err
	}

	totalProducts.ProductDetails = productDetails
	if len(productDetails) == 0 {
		return totalProducts, nil
	}

	totalProducts.TotalCount = productDetails[0].TotalCount
	return totalProducts, nil
}

func UpdateProduct(productID string, productDetails models.ProductUpdateDetails) error {
	SQL := `UPDATE  inventory
            SET     
                    name = $1,
                    price = $2,
                    quantity = $3,
                    updated_at=now()
            WHERE   inventory.id = $4`

	_, err := database.AudiophileDB.Exec(SQL, productDetails.Name, productDetails.Price, productDetails.Quantity, productID)
	if err != nil {
		logrus.Printf("UpdateProduct: cannot update product:%v", err)
		return err
	}
	return nil
}

func DeleteProduct(productID string) error {
	SQL := `UPDATE inventory
            SET    archived_at=now()
            WHERE inventory.id=$1`
	_, err := database.AudiophileDB.Exec(SQL, productID)

	if err != nil {
		logrus.Printf("DeleteProduct: cannot delete product:%v", err)
		return err
	}

	return nil
}

func DeleteProductImage(productImageID string) error {
	SQL := `UPDATE images_per_product
            SET    archived_at=now()
            WHERE images_per_product.id=$1`
	_, err := database.AudiophileDB.Exec(SQL, productImageID)

	if err != nil {
		logrus.Printf("DeleteProductImage: cannot delete product image:%v", err)
		return err
	}

	return nil
}

func AddAddress(userID uuid.UUID, userDetails *models.Users) error {
	SQL := `INSERT INTO user_address(user_id, address)
          VALUES ($1, $2)`
	_, err := database.AudiophileDB.Exec(SQL, userID, userDetails.Address)

	if err != nil {
		logrus.Printf("AddAddress: unable to add address:%v", err)
		return err
	}

	return nil
}

func FetchPrice(productID string) (float64, error) {
	SQL := `SELECT  price
            FROM    inventory
            WHERE   inventory.id=$1
            AND     inventory.archived_at IS NULL `

	var price float64

	err := database.AudiophileDB.Get(&price, SQL, productID)
	if err != nil {
		logrus.Printf("FetchPrice: unable to get price:%v", err)
		return price, err
	}
	return price, nil
}

func AddToCart(productID string, quantity models.Quantity, price float64, userID uuid.UUID) error {
	SQL := `INSERT INTO user_cart_products(product_id, quantity, total_amount, user_id)
            VALUES   ($1, $2, $3, $4)`

	totalPrice := float64(quantity.NumberOfItems) * price

	_, err := database.AudiophileDB.Exec(SQL, productID, quantity, totalPrice, userID)
	if err != nil {
		logrus.Printf("AddToCart: cannot add product to cart:%v", err)
		return err
	}

	return nil
}

func RemoveFromCart(cartID string) error {
	SQL := `UPDATE user_cart_products
            SET    archived_at=now()
            WHERE user_cart_products.id=$1`
	_, err := database.AudiophileDB.Exec(SQL, cartID)

	if err != nil {
		logrus.Printf("RemoveFromCart: cannot remove product from cart:%v", err)
		return err
	}

	return nil
}

func FetchUserAddress(userID uuid.UUID) (uuid.UUID, error) {
	SQL := `SELECT  user_address.id
            FROM    user_address
            WHERE   user_id=$1
            AND     user_address.archived_at IS NULL `

	var addressID uuid.UUID
	err := database.AudiophileDB.Get(&addressID, SQL, userID)
	if err != nil {
		logrus.Printf("FetchUserAddress: unable to get user address:%v", err)
		return addressID, err
	}
	return addressID, nil
}

func FetchTotalAmount(cartID string, userID uuid.UUID) (float64, error) {
	SQL := `SELECT  total_amount
            FROM    user_cart_products
            WHERE   user_cart_products.id = $1
            AND     user_id =$2
            AND     user_cart_products.archived_at IS NULL `
	var totalAmount float64
	err := database.AudiophileDB.Get(&totalAmount, SQL, cartID, userID)
	if err != nil {
		logrus.Printf("FetchTotalAmount: not able to fetch total amount:%v", err)
		return totalAmount, err
	}
	return totalAmount, nil
}

func FetchOrderedProducts(userID uuid.UUID) ([]string, error) {
	var cartID []string

	SQL := `SELECT id
                FROM user_cart_products
                WHERE user_id=$1
                AND   order_check=true
                AND   user_cart_products.archived_at IS NULL `
	err := database.AudiophileDB.Select(&cartID, SQL, userID)
	if err != nil {
		logrus.Printf("FetchOrderedProducts:cannot get product id's:%v", err)
		return cartID, err
	}

	return cartID, nil
}

func SelectProduct(cartID string) error {
	SQL := `UPDATE  user_cart_products
            SET     order_check=true
            WHERE   id=$1`

	_, err := database.AudiophileDB.Exec(SQL, cartID)
	if err != nil {
		logrus.Printf("SelectProducts:cannot select products:%v", err)
		return err
	}
	return nil
}

func RemoveProductsFromCart(userID uuid.UUID, tx *sqlx.Tx) error {
	SQL := `UPDATE user_cart_products
                SET    order_check = false
                WHERE  user_id = $1`
	_, err := tx.Exec(SQL, userID)
	if err != nil {
		logrus.Printf("RemoveProductsFromCart: cannot remove order_check:%v", err)
		return err
	}

	return nil
}

func CheckOut(orderDetails models.OrderDetails) error {
	SQL := `INSERT INTO  order_details(user_id, address_id, total_amount, cart_id)
            VALUES    ($1, $2, $3, $4)`

	_, err := database.AudiophileDB.Exec(SQL, orderDetails.UserID, orderDetails.AddressID, orderDetails.TotalAmount, pq.StringArray(orderDetails.CartID))
	if err != nil {
		logrus.Printf("CheckOut:cannot checkout:%v", err)
		return err
	}
	return nil
}

func InstantPayment(userID uuid.UUID, paymentDetails models.PaymentDetails, orderID string, tx *sqlx.Tx) (uuid.UUID, error) {
	SQL := `INSERT INTO payment(user_id, payment_type, name, account_number, order_id)
            VALUES   ($1, $2, $3, $4, $5)
            RETURNING id`
	var paymentID uuid.UUID
	err := tx.Get(&paymentID, SQL, userID, paymentDetails.PaymentType, paymentDetails.Name, paymentDetails.AccountNumber, orderID)
	if err != nil {
		logrus.Printf("InstantPayment: not able to do payment:%v", err)
		return paymentID, err
	}
	return paymentID, nil
}

func UpdateOrder(orderID string, tx *sqlx.Tx) error {
	SQL := `UPDATE order_details
            SET status=$1,
                updated_at=now()
            WHERE id=$2`
	_, err := tx.Exec(SQL, "completed", orderID)
	if err != nil {
		logrus.Printf("UpdateOrder: cannnot complete payment:%v", err)
		return err
	}
	return nil
}

func CreateBill(userID, paymentID uuid.UUID, orderID string, tx *sqlx.Tx) error {
	SQL := `INSERT INTO bill_details(user_id, payment_id, order_id)
            VALUES   ($1, $2, $3)`
	_, err := tx.Exec(SQL, userID, paymentID, orderID)
	if err != nil {
		logrus.Printf("CreateBill: cannot create bill:%v", err)
		return err
	}
	return nil
}

func ViewBillDetails(userID uuid.UUID) ([]models.BillDetails, error) {
	SQL := `SELECT    user_id,
                      payment_id,
                      order_id
            FROM      bill_details
            WHERE     user_id=$1`

	billDetails := make([]models.BillDetails, 0)

	err := database.AudiophileDB.Select(&billDetails, SQL, userID)
	if err != nil {
		logrus.Printf("ViewBillDetails:%v", err)
		return billDetails, err
	}
	return billDetails, nil
}
