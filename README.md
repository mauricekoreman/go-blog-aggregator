This was a project on Boot.dev

Make sure you have the following installed:
- goose
- sqlc

Im working with postgresql (v14)


in the root of the dir execute `sqlc generate` to generate the database queries into GO code.
Run the migrations with `goose [database connection] up`
