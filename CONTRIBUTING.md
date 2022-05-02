# How to contribute

## Testing

When adding new functionality or implementing a new feature, it would be ideal to have [tests](https://github.com/packethost/ironlib/blob/main/utils/lshw_test.go) and [test_data](https://github.com/packethost/ironlib/blob/main/utils/lshw_test.go#L47) included.

## Submitting changes

Run `make test` and `make lint` and ensure it all passes before submitting a PR.
Checkout the [PR template](https://github.com/packethost/ironlib/blob/main/.github/PULL_REQUEST_TEMPLATE.md) and create a pull request at [ironlib PRs](https://github.com/packethost/ironlib/pulls)

## Coding conventions

https://talks.golang.org/2013/bestpractices.slide#1
https://golang.org/doc/effective_go

Run `make lint` on your changes and that should get it more than half way there.
