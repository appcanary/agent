require 'sinatra'
require 'json'
require 'pry'
require 'base64'

set :bind, '0.0.0.0'
def print_bod(body)
  bod = JSON.load(body.read)
  puts "#" * 10
  if bod && bod["contents"]
    bod["contents"] = Base64.decode64(bod["contents"])
  end
  puts bod
  puts "#" * 10
end

get '/' do 
  "Hello world"
end

post '/api/v1/agent/heartbeat/:id' do
  print_bod(request.body)
  {success: true}.to_json
end

post '/api/v1/agent/servers' do
  print_bod(request.body)
  {uuid:"12345"}.to_json
end

put '/api/v1/agent/servers/:id' do
  print_bod(request.body)
  "OK"
end

get '/api/v1/agent/servers/:id' do
  print_bod(request.body)
  content_type :json
  File.read("dump.json")
end
