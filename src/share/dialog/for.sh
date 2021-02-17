
count=$1
args=$@

echo "args = $args"

for ((i = 2 ; i < $count+2 ; i++)); do
    arg=${!i}
    k=$(i-2)
	echo "$k = $arg"
done

