# Go-Echo Prometheus middleware
Custom Go-Echo web framework Prometheus middleware

## Usage

### With default config
```go
import (
    echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
    echoprom "github.com/andyglass/echo-prometheus"
)

func main() {
	e := echo.New()

	e.Use(echoprom.Middleware())
    e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
    
    e.Logger.Fatal(e.Start(":1323"))
}
```

### With custom config
```go
import (
    echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
    echoprom "github.com/andyglass/echo-prometheus"
)

func main() {
	e := echo.New()

	configMetrics := echoprom.NewConfig()
	configMetrics.Buckets = []float64{
		0.001, // 1ms
		0.005, // 5ms
		0.01,  // 10ms
		0.05,  // 50ms
		0.1,   // 100ms
		0.5,   // 500ms
		1,     // 1s
		2.5,   // 2.5s
		5,     // 5s
		10,    // 10s
	}

	e.Use(echoprom.MiddlewareWithConfig(configMetrics))
    e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
    
    e.Logger.Fatal(e.Start(":1323"))
}
```