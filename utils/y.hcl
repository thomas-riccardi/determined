version = "3.7",
services = {
	api = {
		environment = {
			CONFIG_FILE = "/config/config.json"
		},
		depends_on = [
			"db"
		],
		image = "hashicorpdemoapp/product-api:v0.0.22",
		ports = [
			"19090:9090"
		],
		volumes = [
			"./conf.json:/config/config.json"
		]
	},
	db = {
		image = "hashicorpdemoapp/product-api-db:v0.0.22",
		ports = [
			"15432:5432"
		],
		environment = {
			POSTGRES_PASSWORD = "password",
			POSTGRES_DB = "products",
			POSTGRES_USER = "postgres"
		}
	}
}
