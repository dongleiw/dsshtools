echo -e '127.0.0.1\nlocalhost' > iplist.tmp

function run(){
	echo $@
	./build/dssh $@
	echo 
}
run -H 127.0.0.1 true

run -H 127.0.0.1,localhost true
run -H 127.0.0.1,localhost -H localhost true
run -H 127.0.0.1,localhost -h iplist.tmp true

run -H 127.0.0.1,localhost -h iplist.tmp 'echo $(($RANDOM%2))'
run -g -H 127.0.0.1,localhost -h iplist.tmp 'echo $(($RANDOM%2))'
run -qg -H 127.0.0.1,localhost -h iplist.tmp 'echo $(($RANDOM%2))'

rm iplist.tmp
