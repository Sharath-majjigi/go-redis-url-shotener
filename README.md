# go-redis-url-shotener
A url-shortening application built using GO having Ratelimit and Reset functionality 


How to Build : </br>

1. Clone this repo <pre data-copyable> <div style="background-color: #f7f7f7; padding: 10px;"> git clone https://github.com/Sharath-majjigi/go-redis-url-shotener </div> </pre> 

2. Customise your API_QUOTA and Domain present in .env file </br>

3. Run <pre data-copyable> go run main.go </pre>  or </br> if you want to containerize your application make sure your docker is up and running and hit the command <pre data-copyable> docker-compose up -d </pre> 
   which will create the api and db image into 
   your local Docker registry and once they are up try to access api endpoints
  
