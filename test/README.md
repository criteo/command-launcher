# Integration Test

To run all the integration test simply run the script: `integration.sh`

```bash
cd command-launcher
./test/interation.sh
```

To run particular integration tests, pass the integration test file name in the `integration` folder as the arguments.

For example, the following command runs the tests in:
- test/integration/test-basic.sh
- test/integration/test-exit-code.sh

```bash
cd command-launcher
./test/interation.sh test-basic test-exit-code
```
