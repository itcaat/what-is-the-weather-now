
# Weather Service

This is a simple Go web application that retrieves the user's IP address, determines their location, and fetches the current weather for that location. The application also includes in-memory caching.

## Features

- Determines user's location based on IP. 
- Fetches weather data from wttr.in. 
- Caches IP-to-city mappings for 24 hours. 
- Caches weather responses for 10 minutes. 
- Supports Cloudflare headers for accurate IP detection.
