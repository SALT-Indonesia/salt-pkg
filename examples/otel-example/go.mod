module github.com/SALT-Indonesia/salt-pkg/examples/otel-example

go 1.24.0

replace github.com/SALT-Indonesia/salt-pkg/logmanager => ../../

require (
	github.com/SALT-Indonesia/salt-pkg/logmanager v1.41.0
	github.com/gin-gonic/gin v1.10.1
)
