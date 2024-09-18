-- log_request.lua
local start_time = ngx.req.start_time()
ngx.log(ngx.INFO, "Request started at: ", start_time)
