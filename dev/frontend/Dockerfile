FROM openresty/openresty:alpine

# Copy the Lua script
COPY log_request.lua /etc/nginx/lua/

# Copy index.html and nginx.conf
COPY index.html /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["openresty", "-g", "daemon off;"]
