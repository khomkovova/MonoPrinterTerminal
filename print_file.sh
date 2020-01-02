


id=$(lp $1 | sed 's/request id is //' |  sed 's/ (1 file(s))//')
echo $id
i="0"
while true
do
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
i=$[$i+1]
done