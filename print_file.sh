out=$(lp "$1")
if [ "$?" -ne "0" ]; then
  exit 1
fi

id=$(echo "$out" | sed 's/request id is //' | sed 's/ (1 file(s))//')
i="0"
while true; do
  echo $i
  status=$(lpstat -W not-completed | grep $id)

  if [[ $status == "" ]]; then
    echo "Successful"
    exit 0
  fi

  if [[ $i == 10 ]]; then
    cancel $id
    echo "Timeout"
    exit 1
  fi

  sleep 1
  i=$(($i + 1))
done
