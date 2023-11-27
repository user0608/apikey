## APIKEY Manager

Service responsible for managing the API keys, in conjunction with nginx.

```bash
    server {
        listen 8085;
        server_name pruebaapikey.com;
        error_page 404 = @404;
        location / {
            auth_request /auth;
            proxy_pass http://localhost:8081;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

            error_page 401 = @401;           
        }
        location = /auth {
            internal;
            proxy_pass http://localhost:1323;
            proxy_pass_request_body off;
        }
        location @401 {
            default_type application/json;
            return 401 '{"message":"api key not found or invalid"}';
        }
        location @404 {
            default_type application/json;
            return 404 '{"message":"not found"}';
        }
    }
```

## endpoint

- http://localhost:1323/auth

  - Header: x-api-key

- http://localhost:1323/refresh
- http://localhost:1323/apikeys


### API KEYS
The API keys are stored in a .apikeys file.
