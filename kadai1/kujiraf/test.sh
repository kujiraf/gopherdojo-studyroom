#! /bin/sh

ROOT_PKG=$(cd $(dirname $0); pwd)
TEST_PROFILE_DIR=${ROOT_PKG}/testprofile
IMGCONV=${ROOT_PKG}/imgconv
CONVERTER=${ROOT_PKG}/converter
DIRS=($IMGCONV $CONVERTER)

for dir in ${DIRS[@]}
do
  res=$(find ${dir} -name "*.go" -not -name "*_test.go" | xargs errcheck)
  if [ -n "$res" ]; then
    echo "Missing Error Handling"
    echo "$res"
    exit
  fi
done

# # mainのテスト
go test -v -coverprofile=${TEST_PROFILE_DIR}/imgconv $IMGCONV
go tool cover -html=${TEST_PROFILE_DIR}/imgconv

# # # convertorのテスト
go test -v -coverprofile=${TEST_PROFILE_DIR}/converter $CONVERTER
go tool cover -html=${TEST_PROFILE_DIR}/converter