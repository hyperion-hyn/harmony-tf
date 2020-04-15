# harmony-tf
Harmony TF is a Test Framework for testing various types of test cases and transactions on Harmony's blockchain.

This is a work in progress (WIP) and as such breaking Test Framework API changes can occur at any time until the feature spec/API stabilizes.

## Features

* Automatically importing keys placed in keys/ into the keystore
* Generating temporary receiver accounts
* Sending back any eventual test funds to the originator and subsequently removing the account from the keystore
* Defining test cases and evaluating if a given test case's result matches the expected test result.
* Sending transactions without relying on hmy or any CLI - i.e. directly communicating with the underlying Harmony API:s
