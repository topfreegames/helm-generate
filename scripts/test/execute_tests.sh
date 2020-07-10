#! /bin/bash

# args
COMMAND=$@

# execute tests for each package
failed=0
declare -a failed_packages
for dir in $(go list -f '{{ .Dir }}' ./... | grep -v /vendor/); do
    pushd $dir > /dev/null

    ${COMMAND}
    if [[ "${?}" -ne 0 ]]; then
        failed=1
        failed_packages+=(${dir})
    fi

    popd > /dev/null
done;

# force clean
rm -rf */**/tmp

# final result
if [[ "${failed}" -eq 0 ]]; then
    echo "[SUCCESS] All tests passed successfully!"
    exit 0
else
    echo "[FAIL] Some packages tests have failed: ${failed_packages[@]}"
    exit 1
fi
