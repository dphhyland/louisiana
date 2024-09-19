#!/bin/bash

# Path to the Go binary
GO_PROGRAM="./client"  # or `go run main.go` if running directly with `go run`

# Paths to the JSON files in the /config folder
CONFIG_DIR="./config"
BANK_FILES=("$CONFIG_DIR/bank1_bank_config.json" "$CONFIG_DIR/bank2_bank_config.json" "$CONFIG_DIR/bank3_bank_config.json" "$CONFIG_DIR/bank4_bank_config.json")
TELCO_FILES=("$CONFIG_DIR/telco3_telco_config.json" "$CONFIG_DIR/telco4_telco_config.json")
BIGTECH_FILES=("$CONFIG_DIR/bigtech1_domain_config.json" "$CONFIG_DIR/bigtech1_email_config.json" "$CONFIG_DIR/bigtech1_website_config.json")

# Loop counters
counter=0
bank_modification_counter=0
bank_file_counter=0  # Round-robin counter for bank files

# Function to modify the account name or phone number
modify_value() {
  local value=$1
  local type=$2

  if [[ "$type" == "small" ]]; then
    echo "${value}X"
  elif [[ "$type" == "significant" ]]; then
    echo "$(echo $value | rev)"
  else
    echo "$value"
  fi
}

# ----------------     Function to process a random bank entry
process_random_bank_entry() {
  # Log: Indicate the bank entry is being processed
  echo "Processing bank entry..."

  # Select the next bank file in a round-robin fashion
  bank_file="${BANK_FILES[$bank_file_counter]}"
  bank_file_counter=$(( (bank_file_counter + 1) % ${#BANK_FILES[@]} ))  # Rotate through bank files

  # Log: Print the selected bank file
  echo "Selected bank file: $bank_file"

  # Get the number of entries in the bank file
  array_length=$(jq '.confirmation_of_payee | length' "$bank_file")
  if [[ "$array_length" -eq 0 ]]; then
    echo "Bank file $bank_file is empty or invalid."
    return
  fi

  # Pick a random entry
  random_entry=$(jq -c ".confirmation_of_payee[$((RANDOM % array_length))]" "$bank_file")

  # Extract fields from the JSON entry
  bsb=$(echo "$random_entry" | jq -r '.bsb')
  accountNumber=$(echo "$random_entry" | jq -r '.accountNumber')
  accountName=$(echo "$random_entry" | jq -r '.accountName')

  # Ensure valid data before proceeding
  if [[ -z "$bsb" || -z "$accountNumber" || -z "$accountName" ]]; then
    echo "Invalid bank data in $bank_file, skipping entry."
    return
  fi

  # Decide the modification type for bank account name: Only modify 2 out of every 5 requests
  modify_name=false
  if ((bank_modification_counter % 5 == 0)); then
    modifiedName=$(modify_value "$accountName" "significant")
    modify_name=true
  elif ((bank_modification_counter % 5 == 2)); then
    modifiedName=$(modify_value "$accountName" "small")
    modify_name=true
  else
    modifiedName="$accountName"
  fi

  # Increment the bank modification counter (resets every 5 requests)
  bank_modification_counter=$((bank_modification_counter + 1))

  # Create JSON input for the Go program
  inputJson=$(jq -n --arg bsb "$bsb" --arg accountNumber "$accountNumber" --arg accountName "$modifiedName" \
                '{bank: {bsb: $bsb, accountNumber: $accountNumber, accountName: $accountName}}')

  # Log whether the request was modified
  if [[ "$modify_name" == "true" ]]; then
    echo "Calling the Go program with a modified bank request: $inputJson"
  else
    echo "Calling the Go program with an unmodified bank request: $inputJson"
  fi

  # Call the Go program with the bank details
  $GO_PROGRAM "$inputJson"

  # Add some delay if necessary
  sleep 1
}

# ----------------     Function to process a random telco entry
process_random_telco_entry() {
  # Log: Indicate the telco entry is being processed
  echo "Processing telco entry..."

  # Select a random telco file
  telco_file="${TELCO_FILES[RANDOM % ${#TELCO_FILES[@]}]}"

  # Log: Print the selected telco file
  echo "Selected telco file: $telco_file"

  # Get the number of entries in the telco file
  array_length=$(jq '.threat_metrics | length' "$telco_file")
  if [[ "$array_length" -eq 0 ]]; then
    echo "Telco file $telco_file is empty or invalid."
    return
  fi

  # Pick a random entry
  random_entry=$(jq -c ".threat_metrics | to_entries | .[$((RANDOM % array_length))]" "$telco_file")

  # Extract phone number and metrics
  mobile=$(echo "$random_entry" | jq -r '.key')
  simSwap=$(echo "$random_entry" | jq -r '.value.simSwap')
  fraudReported=$(echo "$random_entry" | jq -r '.value.fraudReported')
  daysActive=$(echo "$random_entry" | jq -r '.value.daysActive')

  # Ensure valid data
  if [[ -z "$mobile" || -z "$simSwap" || -z "$fraudReported" || -z "$daysActive" ]]; then
    echo "Invalid telco data in $telco_file, skipping entry."
    return
  fi

  # Create JSON input for the Go program
  inputJson=$(jq -n --arg mobile "$mobile"  '{telephony: {mobile: $mobile}}')

  # Call the Go program with the modified telco details
  echo "Calling the Go program with modified telco details: $inputJson"
  $GO_PROGRAM "$inputJson"

  # Add some delay if necessary
  sleep 1
}

# ----------------     Function to process a random big tech entry from domain, email, or website data
process_random_bigtech_entry() {
  # Log: Indicate the big tech entry is being processed
  echo "Processing big tech entry..."

  # Select a random big tech file
  bigtech_file="${BIGTECH_FILES[RANDOM % ${#BIGTECH_FILES[@]}]}"

  # Log: Print the selected big tech file
  echo "Selected big tech file: $bigtech_file"

  # Determine the type of the big tech file
  if [[ "$bigtech_file" == *"domain_config"* ]]; then
    process_domain_file "$bigtech_file"
  elif [[ "$bigtech_file" == *"email_config"* ]]; then
    process_email_file "$bigtech_file"
  elif [[ "$bigtech_file" == *"website_config"* ]]; then
    process_website_file "$bigtech_file"
  else
    echo "Unknown big tech file type: $bigtech_file"
  fi
}

# ----------------      Function to process domain file  ---------------------
process_domain_file() {
  local file="$1"

  # Get the number of entries in the domain file
  array_length=$(jq '.domain_threat_metrics | length' "$file")
  if [[ "$array_length" -eq 0 ]]; then
    echo "Domain file $file is empty or invalid."
    return
  fi

  # Pick a random entry
  random_entry=$(jq -c ".domain_threat_metrics | to_entries | .[$((RANDOM % array_length))]" "$file")
  domain=$(echo "$random_entry" | jq -r '.key')
  daysActive=$(echo "$random_entry" | jq -r '.value.daysActive')
  fraudReported=$(echo "$random_entry" | jq -r '.value.fraudReported')
  phishing=$(echo "$random_entry" | jq -r '.value.phishing')
  daysSinceLastUpdate=$(echo "$random_entry" | jq -r '.value.daysSinceLastUpdate')
  # Extract fields from the JSON entry
  domain=$(echo "$random_entry" | jq -r '.key')

  # Ensure valid data
  if [[ -z "$domain" || -z "$daysActive" || -z "$phishing" || -z "$fraudReported" || -z "$daysSinceLastUpdate" ]]; then
    echo "Invalid domain data in $file, skipping entry."
    return
  fi

  # Create JSON input for the Go program
  inputJson=$(jq -n --arg domain "$domain"  \ '{domain: {domain: $domain}}')

  # Call the Go program with the domain details
  echo "Calling the Go program with domain details: $inputJson"
  $GO_PROGRAM "$inputJson"
}

# -------------------- Function to process email file -------------------
process_email_file() {
  local file="$1"

  # Get the number of entries in the email file
  array_length=$(jq '.email_threat_metrics | length' "$file")
  if [[ "$array_length" -eq 0 ]]; then
    echo "Email file $file is empty or invalid."
    return
  fi

  # Pick a random entry
  random_entry=$(jq -c ".email_threat_metrics | to_entries | .[$((RANDOM % array_length))]" "$file")

  # Extract fields from the JSON entry
  email=$(echo "$random_entry" | jq -r '.key')
  daysActive=$(echo "$random_entry" | jq -r '.value.daysActive')
  fraudReported=$(echo "$random_entry" | jq -r '.value.fraudReported')
  spamFlagged=$(echo "$random_entry" | jq -r '.value.spamFlagged')
  daysSinceLastUpdate=$(echo "$random_entry" | jq -r '.value.daysSinceLastUpdate')
  # Ensure valid data
  if [[ -z "$email" || -z "$daysActive" || -z "$fraudReported" || -z "$spamFlagged" || -z "$daysSinceLastUpdate" ]]; then
    echo "Invalid email data in $file, skipping entry."
    return
  fi

  # Create JSON input for the Go program
  inputJson=$(jq -n --arg email "$email" '{emailAddress: {emailAddress: $email}}')

  # Call the Go program with the email details
  echo "Calling the Go program with email details: $inputJson"
  $GO_PROGRAM "$inputJson"
}

# -------------- Function to process website file --------------
process_website_file() {
  local file="$1"

  # Get the number of entries in the website file
  array_length=$(jq '.website_threat_metrics | length' "$file")
  if [[ "$array_length" -eq 0 ]]; then
    echo "Website file $file is empty or invalid."
    return
  fi

  # Pick a random entry
  random_entry=$(jq -c ".website_threat_metrics | to_entries | .[$((RANDOM % array_length))]" "$file")

  # Extract fields from the JSON entry
  website=$(echo "$random_entry" | jq -r '.key')
  daysActive=$(echo "$random_entry" | jq -r '.value.daysActive')
  spamFlagged=$(echo "$random_entry" | jq -r '.value.spamFlagged')
  phishingFlagged=$(echo "$random_entry" | jq -r '.value.phishingFlagged')
  daysSinceLastUpdate=$(echo "$random_entry" | jq -r '.value.daysSinceLastUpdate')

  # Ensure valid data
  if [[ -z "$website" || -z "$daysActive" || -z "$spamFlagged" || -z "$phishingFlagged" || -z "$daysSinceLastUpdate" ]]; then
    echo "Invalid website data in $file, skipping entry."
    return
  fi

  # Create JSON input for the Go program
  inputJson=$(jq -n --arg website "$website" '{website: {website: $website}}')

  # Call the Go program with the website details
  echo "Calling the Go program with website details: $inputJson"
  $GO_PROGRAM "$inputJson"
}

#----------------      Main function to run randomly between bank, telco, and big tech entries continuously
run_continuously() {
  while true; do
    counter=$((counter+1))

    # Randomly choose between bank, telco, or big tech operation
    case $((RANDOM % 3)) in
      0) process_random_bank_entry ;;
      1) process_random_telco_entry ;;
      2) process_random_bigtech_entry ;;
    esac

    # Add delay between requests if needed
    sleep 1
  done
}

# Start the continuous execution
run_continuously