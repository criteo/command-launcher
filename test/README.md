# Integration Test

To run all the integration test simply run the script: `integration.sh`

```bash
cd command-launcher
./test/integration.sh
```

To run particular integration tests, pass the integration test file name (without .sh) in the `integration` folder as the arguments.

For example, the following command runs the tests in:
- test/integration/test-basic.sh
- test/integration/test-exit-code.sh

```bash
cd command-launcher
./test/integration.sh test-basic test-exit-code
```

Copy the test-template.sh to create a new test suite
