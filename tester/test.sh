#!/bin/bash

HOST=http://localhost:8081

get_token() {
	auth_reponse=$(curl --location "$HOST/auth/sessions" \
		--silent \
		--header 'Content-Type: application/json' \
		--data-raw '{"user":{"email":"opg_document_and_d@publicguardian.gsi.gov.uk","password":"$2y$10$Xlq5mrdU6ZSh7kU5Yi.vpuCOrWCekNl9BwLcAg5G5bwr22ehTEpEa"}}')

	echo $auth_reponse | jq -r '.authentication_token'
}

xml_to_set() {
	type=$1
	xml=$2
	caseno=$3

	xml_b64=$(echo $xml | base64)

	echo "<?xml version=\"1.0\" encoding=\"UTF-8\"?>
<Set xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xsi:noNamespaceSchemaLocation=\"SET.xsd\">
<Header CaseNo=\"$caseno\" Scanner=\"9\" ScanTime=\"2014-09-26 12:38:53\" ScannerOperator=\"Administrator\" Schedule=\"01-0001253-20160909174150\" />
<Body>
<Document Type=\"$type\" Encoding=\"UTF-8\" NoPages=\"1\">
<XML>$xml_b64</XML>
<PDF>SUkqAAoAAAAAtBEAAAEDAAEAAAABAAAAAQEDAAEAAAABAAAAAgEDAAEAAAABAAAAAwEDAAEAAAABAAAABgEDAAEAAAABAAAACgEDAAEAAAABAAAADQECAAEAAAAAAAAAEQEEAAEAAAAIAAAAEgEDAAEAAAABAAAAFQEDAAEAAAABAAAAFgEDAAEAAAAAIAAAFwEEAAEAAAABAAAAGgEFAAEAAADcAAAAGwEFAAEAAADkAAAAHAEDAAEAAAABAAAAKAEDAAEAAAACAAAAKQEDAAIAAAAAAAEAAAAAAEgAAAABAAAASAAAAAEAAAA=</PDF>
</Document>
</Body>
</Set>"
}

upload() {
	token=$1
	type=$2
	file_name=$3
	include_caseno=$4

	if [[ "$include_caseno" == "Y" ]]; then
		caseno="7000-1234-1234"
	else
		caseno=""
	fi

	xml=$(cat "./service-app/xml/$file_name")
	set=$(xml_to_set $type "$xml" "$caseno")

	response=$(curl --location "$HOST/api/ddc" \
		--silent \
		--header 'Content-Type: text/xml' \
		--header "Cookie: membrane=$token" \
		--data "$set")

	echo $response
}

check_file() {
	token=$1
	type=$2
	file=$3

	docker compose exec localstack awslocal s3 rm s3://opg-backoffice-jobsqueue-local --recursive --quiet
	docker compose exec localstack awslocal sqs purge-queue --queue-url=ddc.fifo

	upload_response=$(upload $token $type "$file.xml")
	upload_success=$(echo $upload_response | jq -r .data.success)
	if [[ ! $upload_success = "true" ]]; then
		echo -e "\033[31m$file failed: upload failed with error $(echo $upload_response | jq .data.message)\033[0m"
		echo -e "\033[31mValidation errors: $(echo $upload_response | jq .data.validationErrors)\033[0m"
		return 1
	fi

	upload_uid=$(echo $upload_response | jq -r .data.uid)

	message_body=$(docker compose exec localstack awslocal sqs receive-message --queue-url=ddc.fifo | jq ".Messages[0].Body")

	if [[ ! $message_body =~ "$upload_uid" ]]; then
		echo -e "\033[31m$file failed: message body doesn't contain $upload_uid (is: $message_body)\033[0m"
		return 1
	fi

	s3_files=$(docker compose exec localstack awslocal s3api list-objects --bucket=opg-backoffice-jobsqueue-local | jq '.Contents | length')
	if [[ ! $s3_files = "2" ]]; then
		echo -e "\033[31m$file failed: s3 should contain 2 files, but contains $s3_files\033[0m"
		return 1
	fi

	echo -e "\033[32m$file passed\033[0m"
	return 0
}

check_attachment() {
	token=$1
	type=$2
	file=$3

	docker compose exec localstack awslocal s3 rm s3://opg-backoffice-jobsqueue-local --recursive --quiet

	upload_response=$(upload $token $type "$file.xml" "Y")
	upload_success=$(echo $upload_response | jq -r .data.success)
	if [[ ! $upload_success = "true" ]]; then
		echo -e "\033[31m$file failed: upload failed with error $(echo $upload_response | jq .data.message)\033[0m"
		echo -e "\033[31mValidation errors: $(echo $upload_response | jq -r .data.validationErrors)\033[0m"
		return 1
	fi

	s3_files=$(docker compose exec localstack awslocal s3api list-objects --bucket=opg-backoffice-jobsqueue-local | jq '.Contents | length')
	if [[ ! $s3_files = "2" ]]; then
		echo -e "\033[31m$file failed: s3 should contain 2 files, but contains $s3_files\033[0m"
		return 1
	fi

	echo -e "\033[32m$file passed\033[0m"
	return 0
}

do_test() {
	token=$(get_token)

	check_file $token "EP2PG" "EP2PG-valid"
	check_file $token "LP1F" "LP1F-valid"
	check_file $token "LP1H" "LP1H-valid"
	check_file $token "LP2" "LP2-valid"

	check_attachment $token "Correspondence" "Correspondence-valid"
	check_attachment $token "EPA" "EPA-valid"
	check_attachment $token "LPA002" "LPA002-valid"
	check_attachment $token "LPA114" "LPA114-valid"
	check_attachment $token "LPA115" "LPA115-valid"
	check_attachment $token "LPA116" "LPA116-valid"
	check_attachment $token "LPA117" "LPA117-valid"
	check_attachment $token "LPA120" "LPA120-valid"
	check_attachment $token "LPA-PA" "LPA-PA-valid"
	check_attachment $token "LPA-PW" "LPA-PW-valid"
	check_attachment $token "LPC" "LPC-valid"
}

result=$(do_test)
echo "$result"
if [[ $result =~ "failed" ]]; then
	exit 1
fi
