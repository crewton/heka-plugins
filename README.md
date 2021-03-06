# heka-plugins

## Installation
Follow the [guide from the heka documentation to build with external plugins][3].

You also need to include the official aws-sdk-go by adding the following below the `goamz` clone in `cmake/externals.cmake`:
```bash
git_clone(https://github.com/vaughan0/go-ini a98ad7ee00ec53921f08832bc06ecf7fd600e6a1)
git_clone(https://github.com/aws/aws-sdk-go 90a21481e4509c85ee68b908c72fe4b024311447)
add_dependencies(aws-sdk-go go-ini)
```

If you do not need all of the plugins from this repository, you can specify specific ones:
```bash
add_external_plugin(git https://github.com/crewton/heka-plugins master kinesis)
```

## kinesis
This output will put your [heka][1] messages and put them into a [Kinesis][2] stream.

Example configuration:

```ini
[KinesisOut]
type = "KinesisOutput"
region = "us-east-1"
stream = "foobar"
access_key_id = "AKIAJ89854WHHJDF8HJF"
secret_access_key = "JKLjkldfjklsdfjkls+d8u8954hjkdfkfdfgfj"
payload_only = false
message_matcher = "TRUE"
```

  [1]: https://hekad.readthedocs.org/en/latest/index.html
  [2]: https://aws.amazon.com/kinesis/
  [3]: http://hekad.readthedocs.org/en/latest/installing.html#building-hekad-with-external-plugins