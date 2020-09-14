# pb

## Usage

``` bash
$ pb -F example/helloworld/helloworld.proto ls messages
helloworld.HelloRequest
helloworld.HelloReply

$ echo '{"name": "ktr"}' | pb -F example/helloworld/helloworld.proto decode helloworld.HelloRequest
{
  "name": "ktr"
}
```
