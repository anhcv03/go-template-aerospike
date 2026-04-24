#!/bin/bash
if [[ $EUID -eq 0 ]]; then
   echo "This script must not be run as root" 
   exit 1
fi

# echo "Check last video file for metadata error"

# TODO: If there are error videos, we have to upload it
# FIXED_VIDEO_FILE=$(python check_video_error.py)
# echo "Finish fixing an error video $FIXED_VIDEO_FILE"
#VIDEO FILE LENGTH (seconds)
VIDEO_DURATION=300.0

# Number of days videos will be kept in local storage
KEEP_VIDEO_DAYS=1
#VIDEO FILE STORAGE LOCATION
VIDEO_DIR=$(echo -e $HOME)/bms-videos

echo "Video dir $VIDEO_DIR"

# sleep 10
#CLIENT IP SERVICE
SERVER_CLIENTIP=127.0.0.1:19080
#AFAD VIDEO SERVICE
SERVER_VIDEO=127.0.0.1:19080
# API_GET_CLIENT_IP=http://$SERVER_CLIENTIP/ip
API_UPLOAD_VIDEO=http://$SERVER_VIDEO/videos/upload
API_UPLOAD_FILE=http://$SERVER_VIDEO/files
API_REGISTER_CLIENT_IP=http://$SERVER_VIDEO/clients
API_VIDEO_CREATE=http://$SERVER_VIDEO/videos

getCurrentTimeStamp() {
    echo `date +%s`;
}

getClientIP() {
	echo "Try to get client with API $API_GET_CLIENT_IP"
	CLIENT_IP=$(curl -X POST --write-out %{http_code} "$API_REGISTER_CLIENT_IP" | jq -r '.ip?')
    echo "Get client ip result $http_code --> $CLIENT_IP"
}

registerClientIP() {
    # getClientIP
    echo "Try to register client ip $CLIENT_IP with API $API_REGISTER_CLIENT_IP"
    REGISTER_IP=$(curl -X POST --write-out %{http_code} "$API_REGISTER_CLIENT_IP" | jq -r '.ip?')
    echo "Register client result $http_code --> $REGISTER_IP"
}

uploadVideo() {
    
    END_TIMESTAMP=$(getCurrentTimeStamp)
    VIDEO_RECORDED_DURATION=`expr $END_TIMESTAMP - $START_TIMESTAMP`
    echo "video file $VIDEO_FILE"
    if [ -f "$VIDEO_FILE" ] && [ $VIDEO_RECORDED_DURATION -gt 10 ]
    then
        UPLOAD_FILE=$VIDEO_FILE
        upload &    
    else
        echo "Cannot upload video $VIDEO_FILE of client $CLIENT_IP, time $START_TIMESTAMP - $END_TIMESTAMP since file is not existed, or video duration is too short"
    fi
}

upload() {
    getClientIP
    # echo "Try to upload video with API $API_UPLOAD_VIDEO"
	# echo "Upload video file $UPLOAD_FILE of client $CLIENT_IP, time $START_TIMESTAMP - $END_TIMESTAMP to $SERVER_IP at $(date "+%Y%m%d_%H%M%S")"
	# UPLOAD_RESULT=$(curl -X POST -F "upload=@$UPLOAD_FILE" "$API_UPLOAD_VIDEO?client_ip=$CLIENT_IP&start_time=$START_TIMESTAMP&end_time=$END_TIMESTAMP")
    file_id=$(curl -XPOST -F "file=@$UPLOAD_FILE" $API_UPLOAD_FILE | jq -r '.id?')
    echo $file_id
    video_info=$(jq -n "{ip: \"$CLIENT_IP\", start_time: $START_TIMESTAMP, end_time: $END_TIMESTAMP, saved: false, file_id: \"$file_id\"}")
    echo $video_info
    curl -XPOST $API_VIDEO_CREATE -H 'Content-Type: application/json' -d "$video_info"

}

cleanOutOfDateVideo() {
    echo "Clean out-of-date videos"
    find $VIDEO_DIR -mtime +$KEEP_VIDEO_DAYS -type f -delete 1>/dev/null 2>&1 &
}

echo "Check video folder!"
if [ -d $VIDEO_DIR ];
then
    echo "Video folder is already existed!"
else
    mkdir $VIDEO_DIR
    echo "Create folder for video!"
fi
registerClientIP

while true
do
  START_TIME=$(date "+%Y%m%d_%H%M%S")
  START_TIMESTAMP=$(getCurrentTimeStamp)

  VIDEO_FILE=$VIDEO_DIR/$START_TIME.mp4
  SCREEN_RESOLUTION=$(xdpyinfo | awk '/dimensions:/ { print $2; exit }')
  echo "Capture to file $VIDEO_FILE, screen resolution $SCREEN_RESOLUTION"
  
  #/usr/bin/ffmpeg -s 3840x1080 -video_size 3840x1080 -f x11grab -threads 0 -framerate 5 -i :0.0+0,0 -c:v libx264 -preset ultrafast -crf 30 -y "$VIDEO_FILE" 1>/dev/null 2>&1 &
  /usr/bin/ffmpeg -s $SCREEN_RESOLUTION -video_size $SCREEN_RESOLUTION -f x11grab -threads 0 -framerate 5 -t $VIDEO_DURATION -i $DISPLAY -c:v libx264 -preset slow -crf 30 -y "$VIDEO_FILE"

  uploadVideo
  cleanOutOfDateVideo
  # Sleep one second to prevent upload too fast
  sleep 30
done
