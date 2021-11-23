#!/usr/bin/env bash

# This script performs the following:
# 1. Run the query, use jq to capture the QueryExecutionId, and then capture that into bash variable
# 2. Wait for the query to finish running (240 seconds).
# 3. Get the results.
# 4. Json data points struct build

# Expected env variables are:
# AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY for accessing AWS resources
# AWS_ATHENA_SOURCE_TABLE
# AWS_ATHENA_OUTPUT_LOCATION
# GITHUB_REPOSITORY

set -euo pipefail

! read -r -d '' query << EOM
select
replace(url_extract_path("d.url"), '/arduino-cli/arduino-cli_', '') as flavor,
count("id") as gauge
from ${AWS_ATHENA_SOURCE_TABLE}
where "d.url" like 'https://downloads.arduino.cc/arduino-cli/arduino-cli_%'
and "d.url" not like '%latest%' -- exclude latest redirect
and "d.url" not like '%alpha%' -- exclude early alpha releases
and "d.url" not like '%.tar.bz2%' -- exclude very old releases archive formats
group by 1
EOM

queryExecutionId=$(
aws athena start-query-execution \
--query-string "${query}" \
--query-execution-context "Database=demo_books" \
--result-configuration "OutputLocation=${AWS_ATHENA_OUTPUT_LOCATION}" \
--region us-east-1 | jq -r ".QueryExecutionId"
)

echo "QueryExecutionId is ${queryExecutionId}"
for i in $(seq 1 120); do
  queryState=$( aws athena get-query-execution \
  --query-execution-id "${queryExecutionId}"  \
  --region us-east-1 | jq -r ".QueryExecution.Status.State"
  );

  if [[ "${queryState}" == "SUCCEEDED" ]]; then
      break;
  fi;

  echo "QueryExecutionId ${queryExecutionId} - state is ${queryState}"

  if [[ "${queryState}" == "FAILED" ]]; then
      exit 1;
  fi;

  sleep 2
done

echo "Query succeeded. Processing data"
queryResult=$( aws athena get-query-results \
--query-execution-id "${queryExecutionId}" \
--region us-east-1 | jq --compact-output
);

! read -r -d '' jsonTemplate << EOM
{
"type": "gauge",
"name": "arduino.downloads.total",
"value": "%s",
"host": "${GITHUB_REPOSITORY}",
"tags": [
"version:%s",
"os:%s",
"arch:%s",
"cdn:downloads.arduino.cc",
"project:arduino-cli"
]
},
EOM

datapoints="["
for row in $(echo "${queryResult}" | jq 'del(.ResultSet.Rows[0])' | jq -r '.ResultSet.Rows[] | .Data' --compact-output); do
  value=$(jq -r ".[1].VarCharValue" <<< "${row}")
  tag=$(jq -r ".[0].VarCharValue" <<< "${row}")
  # Some splitting to obtain 0.6.0, Windows, 32bit elements from string 0.6.0_Windows_32bit.zip
  split=($(echo "$tag" | tr '_' '\n'))
  if [[ ${#split[@]} -ne 3 ]]; then
    continue
  fi
  archSplit=($(echo "${split[2]}" | tr '.' '\n'))
  datapoints+=$(printf "${jsonTemplate}" "${value}" "${split[0]}" "${split[1]}" "${archSplit[0]}")
done
datapoints="${datapoints::-1}]"

echo "::set-output name=result::$(jq --compact-output <<< "${datapoints}")"
