import yt_dlp

def download_audio(youtube_link):
    # Create a yt_dlp options object
    ydl_opts = {
        'verbose': False,
        'format': 'bestaudio/best',
        'postprocessors': [{
            'key': 'FFmpegExtractAudio',
            'preferredcodec': 'mp3',
            'preferredquality': '192',
        }],
        'outtmpl': './audio/%(title)s.%(ext)s',  # Output template for the filename
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        ydl.download([youtube_link])

def scrape_links(filename):
    file = open(filename, "r")
    lines = file.readlines()

    for line in lines:
        try:
            download_audio(line)
        except:
            print(f'could not download: {line}\n')


scrape_links("links.txt")
