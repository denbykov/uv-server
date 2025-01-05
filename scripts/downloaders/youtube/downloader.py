import os
from sys import argv
from json import dumps
from argparse import ArgumentParser

import yt_dlp

DOWNLOADING_PROGRESS = 1
DOWNLOADING_DONE = 2
DOWNLOADING_FAILED = 3

class Logger:
    def debug(self, msg):
        pass

    def info(self, msg):
        pass

    @staticmethod
    def error(msg):
        print(msg)


def download_file(url: str, dir: str, ffmpeg_location: str):
    filename = str()

    def progress_hook(data):
        if data['status'] == 'finished':
            nonlocal filename
            filename = data['filename']
        else:
            progress = {
                "type": DOWNLOADING_PROGRESS,
                "percentage": data["downloaded_bytes"] / data["total_bytes"] * 100
            }
            
            # data["speed"]
            
            print(dumps(progress))

    ydl_opts = {
        "format": "bestaudio/best",
        'outtmpl': f'{dir}/%(title)s.%(ext)s',
        "postprocessors": [
            {
                "key": "FFmpegExtractAudio",
                'preferredcodec': 'mp3',
            }
        ],
        "ffmpeg_location": ffmpeg_location,
        'logger': Logger(),
        'progress_hooks': [progress_hook],
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        error_code = ydl.download(url)
    
    return filename[:filename.rfind('.')] + '.mp3'

if __name__ == "__main__":
    parser = ArgumentParser()

    parser.add_argument(
        "--url", type=str, nargs=1, required=True,
        help="Url to the file to be downloaded")
    
    parser.add_argument(
        "--dir", type=str, nargs=1, required=True,
        help="Directory to store the file")
    
    parser.add_argument(
        "--ffmpeg_location", type=str, nargs=1, required=True,
        help="ffmpeg location")
    
    namespace = parser.parse_args(argv[1:])

    filename = download_file(
        namespace.url[0],
        namespace.dir[0],
        namespace.ffmpeg_location[0])
    
    progress = {
        "type": DOWNLOADING_DONE,
        "filename": filename
    }
    
    print(dumps(progress))
