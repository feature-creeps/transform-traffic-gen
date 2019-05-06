# transform-traffic-gen

Finds images with the name field prefixed with `gen_up_` and then creates transformation requests to be fed into [vegeta](https://github.com/tsenart/vegeta)

Use the latest release of the [uploader](https://github.com/feature-creeps/upload-traffic-gen) so images are set with the prefix.

You'll need go 1.12.x installed - run

```console
go run main.go | \
  vegeta attack -rate=1/s -lazy -format=json -duration=10s | \
  vegeta report
```
