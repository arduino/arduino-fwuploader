path=$1 # the path of the directory where the files and directories that need to be copied are located
s3Dir=$2 # the s3 bucket path

for entry in "$path"/*; do
    name=`echo $entry | sed 's/.*\///'`  # getting the name of the file or directory
    if [[ -d  $entry ]]; then  # if it is a directory
        aws s3 cp  --recursive "$name" "$s3Dir/$name/"
    else  # if it is a file
        aws s3 cp "$name" "$s3Dir/" --exclude "generator.py" --exclude "raw_boards.json" --exclude "s3Copy.sh"
    fi
done
