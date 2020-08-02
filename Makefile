n=9999
default:
	-/bin/rm -rf tmp
	time go run main.go -h -t -n "$(n)"
	mencoder mf://tmp/*.png -mf w=1280:h=720:fps=30:type=png -ovc copy -oac copy -o film.avi 
	-/bin/rm -f film.mp4
	time ffmpeg  -i film.avi -i basicly-1511.wav  film.mp4
	mplayer -fs film.mp4              # preview
	echo mplayer -fs -loop 0 film.mp4 # suggestion

clean:
	-/bin/rm -rf tmp film.avi film.mp4
