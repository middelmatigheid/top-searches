local searches = {"phone", "pants", "books", "backpack", "soap", "laptop", "mouse", "keyboard", "monitor", "chair", "knife", "toys"}
local users = {"alice", "bob", "charlie", "diana", "eve", "frank", "grace", "henry", "iris", "jack", "tom", "anna", "joe", "leon", "mark", "donald", "chris"}

request = function()
    local search = searches[math.random(#searches)]
    local user = users[math.random(#users)]
    
    local body = string.format('{"search":"%s","user":"%s"}', search, user)
    
    return wrk.format("POST", nil, {["Content-Type"] = "application/json"}, body)
end