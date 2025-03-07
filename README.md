This was a project on Boot.dev, making a blog aggregator

Make sure you have the following installed:
- Golang
- Postgresql (im using v14)
- goose
- sqlc

Setup a config file named `.gatorconfig.json` in your HOME directory with the following content: 
```
{"db_url":"postgres://{db_username}:{db_password}@localhost:5432/gator?sslmode=disable","current_user_name":""}
```

in the root of the dir execute `sqlc generate` to generate the database queries into GO code.
Run the migrations with `goose [database connection] up`

To install the blog-aggregator to run anywhere use: `go install`