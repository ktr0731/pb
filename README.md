# pb

## Usage

``` bash
$ pb -F example/helloworld/helloworld.proto ls messages
helloworld.HelloRequest
helloworld.HelloReply

$ echo 'CgNrdHI=' | pb -F example/helloworld/helloworld.proto decode --in base64 helloworld.HelloRequest
{
  "name": "ktr"
}
```
