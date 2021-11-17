# nordigen-lunchmoney-sync

Sync transactions from bank accounts to Lunchmoney via the Nordigen API.

## Configuration

Execute the script with environment variables to sync transactions:
```
NORDIGEN_SECRET_ID=[Nordigen Secret ID from the Nordigen Control Panel]
NORDIGEN_SECRET_KEY=[Nordigen Secret Key from the Nordigen Control Panel]

LUNCHMONEY_ACCESS_TOKEN=[Lunchmoney Access Token from the Lunchmoney App]

# configure the mapping of Nordigen Accoutn IDs to Lunchmoney Asset IDs (Lunchmoney Accounts)
MAPPING="[Nordigen Account ID]:[Lunchmoney Asset ID],[Nordigen Account ID]:[Lunchmoney Asset ID][,…]"
```

If all parameters are specified the script will sync the transactions and then exit.

## Get the Nordigen Account ID

Currently the Nordigen Account set up is not automated and has to be done by hand via the HTTP API. To get started you need to create a [Nordigen Account](https://nordigen.com/).

You will need to create a new pair of User Secrets in the [Nordigen Control Panel](https://ob.nordigen.com/user-secrets/). Note down the Secret ID and Secret Key.

The instructions are basically a simplified version of the official [Nordigen Quickstart Guide](https://nordigen.com/en/account_information_documenation/integration/quickstart_guide/). Please check this guide in case there are any issues or questions.

Create an access token using the Secret ID and Secret Key.
```
curl -X "POST" "https://ob.nordigen.com/api/v2/token/new/" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "secret_id": "[Nordigen Secret ID]",
  "secret_key": "[Nordigen Secret Key]"
}'
```
Note down the Nordigen Access Token from the `access` field from the response.

Fetch a list of all available institutions. You may want to change the country query parameter.
```
curl "https://ob.nordigen.com/api/v2/institutions/?country=de" \
     -H 'Authorization: Bearer [Nordigen Access Token]'
```

Look for the institution you want to connect and not down its ID. For this example we will go with PayPal so the ID is `PAYPAL_PPLXLULL`.

Build a link to authenticate the institution. `reference` can be an arbitrary string unique for the institution. `redirect` URL does not matter.
```
curl -X "POST" "https://ob.nordigen.com/api/v2/requisitions/" \
     -H 'Authorization: Bearer [Nordigen Access Token]' \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "user_language": "EN",
  "redirect": "http://127.0.0.1",
  "reference": "PAYPAL",
  "institution_id": "PAYPAL_PPLXLULL"
}'
```

Note down the `id` field from the response, this is the Nordigen Requisition ID. Visit the link from the `link` field from the response and authenticate your bank account.

Now you are done connecting your account to Nordigen. A Nordigen Requisition ID basically represents a connection to a bank user. Each user may have multiple accounts. You can run this script to view all accounts for a Requisition ID. Execute the script with the following environment variables.
```
NORDIGEN_SECRET_ID=[Nordigen Secret ID from the Nordigen Control Panel]
NORDIGEN_SECRET_KEY=[Nordigen Secret Key from the Nordigen Control Panel]

LUNCHMONEY_ACCESS_TOKEN=[Lunchmoney Access Token from the Lunchmoney App]

NORDIGEN_REQUISITION_IDS=[Nordigen Requisition ID]
```

It will print all Nordigen Account IDs and Lunchmoney Asset IDs so you can create the right mapping. You will want to find the right "nordigen account" message and copy the ID from there. Next you find the right "lunchmoney account" message and copy its ID as well. You can create a mapping like this: 
```
MAPPING="[Nordigen Account ID]:[Lunchmoney Asset/Acccount ID]"
```
Multiple mappings can be seperated via commas. If you run the script with the `MAPPING` variable set it will sync the transactions and then exit.
