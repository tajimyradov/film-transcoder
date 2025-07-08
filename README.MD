# Film & TV Series Transcoder

This application is used to transcode Videos to the HLS (HTTP Live Streaming) format.  
It supports multiple audio languages and subtitles.

---

## 🚀 Quick Start

**1. Prepare Your Source:**  
Use [HandBrake](https://handbrake.fr/) to set handler names for audio tracks in the following format:

- `1-0-eng` &nbsp;→&nbsp; Original audio in English  
- `2-0-rus` &nbsp;→&nbsp; Dubbing in Russian  
- `3-lf-rus` &nbsp;→&nbsp; Studio dubbing (Lostfilm, Russian)

Install [FFMPEG](https://ffmpeg.org/) and [FFPROBE](https://ffmpeg.org/ffprobe.html)

**2. Handler Name Format:**  
`<type>-<studio>-<language>`

| Type | Description      |
|------|-----------------|
| 1    | Original        |
| 2    | Dubbing         |
| 3    | Studio Dubbing  |

- **Example:**  
  - `1-0-eng` → Original English audio  
  - `2-0-rus` → Dubbing in Russian  
  - `3-lf-rus` → Studio (Lostfilm) dubbing in Russian

---

## 📚 Handler Name Details

- **First value:**  
  - `1` → Original  
  - `2` → Dubbing  
  - `3` → Studio

- **Second value:**  
  - Studio code (e.g., `lf` for Lostfilm, or `0` if not applicable)

- **Third value:**  
  - ISO 639-2 language code (e.g., `eng`, `rus`)

---

## 📝 Example Definitions

- `1-0-eng`: Original voice of the film in English.
- `2-0-rus`: Dubbing of the film in Russian.
- `3-lf-rus`: Russian audio made by Lostfilm studio.

---

Enjoy seamless transcoding for multilingual films and series!