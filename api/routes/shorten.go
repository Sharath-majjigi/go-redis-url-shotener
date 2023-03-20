package routes

import (
	"os"
	"sharath/database"
	"sharath/helpers"
	"strconv"
	"time"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
    URL             string              `json: "url"`
	CustomShort     string              `json: "short"`
	Expiry          time.Duration       `json: "expiry"`     
}

type response struct {
    URL               string             `json: "url"`
	CustomShort		  string             `json: "short"`
	Expiry			  time.Duration      `json: "expiry"` 
	XRateRemaining    int                `json: "rate_limit"` 
	XRateLimitReset   time.Duration      `json: "rate_limit_reset"` 
}

func ShortenURL(c *fiber.Ctx) error{

	body := new(request)

	if err := c.BodyParser(&body); err !=nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Cannot parse JSON"})
	}

	//Rate Limiting

	r2 := database.CreateClient(1)
	defer r2.Close()

	value , err :=	r2.Get(database.Ctx,c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx,c.IP(),os.Getenv("API_QUOTA"),30*60*time.Second).Err()
	} else {
		// value, _ := r2.Get(database.Ctx,c.IP()).Result()
		valInt, _ := strconv.Atoi(value)
		if valInt <= 0{
			limit, _ := r2.TTL(database.Ctx,c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"err" : "Rate limit Exceeded",
			"rate_limit_reset" : limit / time.Nanosecond / time.Minute ,})
		}
	}



	//Checking if users request is actual URL 
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Not A Valid URL"})
	}

	
	//Checking for domain errors
	if !helpers.RemoveDomainError(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"You cannot access this :("})
	}

	//Enforcing SSL (HTTPS)

	body.URL=helpers.EnforceHttp(body.URL)

	var id string

	if body.CustomShort == ""{
		id = uuid.New().String()[:6]
	}else {
		id=body.CustomShort
	}

	r := database.CreateClient(0)
    defer r.Close()

	val, _ := r.Get(database.Ctx,id).Result()

	if val != ""{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error" : "URL customShort is Already in use :( "})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}


	err = r.Set(database.Ctx,id,body.URL,body.Expiry*3600*time.Second).Err()

	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"unable to connect to server !"})
	}


	resp := response{
		URL: body.URL,
		CustomShort: "",
		Expiry: body.Expiry,
		XRateRemaining: 10,
		XRateLimitReset:30,
	}

	r2.Decr(database.Ctx,c.IP())

	val, _ = r2.Get(database.Ctx,c.IP()).Result()
	resp.XRateRemaining, _ =strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx,c.IP()).Result()
	resp.XRateLimitReset = ttl /time.Nanosecond /time.Minute


	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}