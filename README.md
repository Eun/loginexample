**loginexample** is an example of [go-gen-api](https://github.com/Eun/go-gen-api)

## Details

    /_gogenapi           Contains a build.go file that runs go-gen-api to generate the api
    /cmd/loginexample    The main application

## Run the example

### Visual Studio Code
If you have Visual Studio Code you could just use `Start Debugging` (<kbd>F5</kbd>) the api will be autogenerated to `/gogenapi`

### Manual

Generate the api:

    go run _gogenapi/build.go

Run the program:

    go run github.com/Eun/loginexample/cmd/loginexample

## Usage
An http server will be started on `:8000` you can use following REST API calls:

| URL             | Method   | Body                                                           | Headers                                                                | Comments                                                 |
|-----------------|----------|----------------------------------------------------------------|------------------------------------------------------------------------|----------------------------------------------------------|
| `/user/create`  | `POST`   | `{ "Name": "Alice", "Password: "password" }`                   | `Content-Type: application/json`                                       |                                                          |
| `/user/login`   | `POST`   | `{ "Name": "Alice", "Password: "password" }`                   | `Content-Type: application/json`                                       |                                                          |
| `/user/logout`  | `GET`    |                                                                | `Token: <Token From Login Response>`                                   |                                                          |
| `/user/get`     | `GET`    |                                                                | `Token: <Token From Login Response>`                                   |                                                          |
| `/user/delete`  | `GET`    |                                                                | `Token: <Token From Login Response>`                                   |                                                          |
| `/user/update`  | `POST`   | `{ "Update": { "Name": "Bob" } }`                              | `Content-Type: application/json`, `Token: <Token From Login Response>` | Update only the fields you want to change                |
| `/admin/get`    | `POST`   | `{ "Name": "Alice"}`                                           | `Content-Type: application/json`                                       | Body is optional, you can filter any field the table has |
| `/admin/delete` | `POST`   | `{ "Name": "Alice"}`                                           | `Content-Type: application/json`                                       | Body is optional, you can filter any field the table has |
| `/admin/update` | `POST`   | `{ "Find": { "Name": "Alice" }, "Update": { "Name": "Bob" } }` | `Content-Type: application/json`                                       | Body is optional, you can filter any field the table has |

> Note that the admin functions do not check for authentication