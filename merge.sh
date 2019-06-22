branch=$(git branch | grep \* | cut -d ' ' -f2)
num=$(echo "${branch: -1}")
echo $num

while [ $num -le 6 ]
do 
    git checkout c$((num+1))
    git merge --no-edit c$num
    num=$((num+1))
    echo $num
done 
git checkout $branch