#!/bin/bash
mkdir -p assets/audio

# Voices: 
# Orc: Ralph (Deep)
# Demon: Whisper
# Peasant: Alex (Normal)

# Orc Menace
say -v Ralph "I will crush you!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/orc_menace_1.wav
say -v Ralph "For the horde!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/orc_menace_2.wav
say -v Ralph "Fresh meat!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/orc_menace_3.wav

# Demon Menace
say -v Whisper "Your soul is mine." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/demon_menace_1.wav
say -v Whisper "Burn in the abyss." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/demon_menace_2.wav
say -v Whisper "The end is near." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/demon_menace_3.wav

# Peasant Menace
say -v Alex "Get off my land!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/peasant_menace_1.wav
say -v Alex "I've had enough of you!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/peasant_menace_2.wav
say -v Alex "Take this!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/peasant_menace_3.wav

# Agonic Death Cries
say -v Ralph "NOOOooooooo..." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/orc_death.wav
say -v Whisper "I return... to the... shadow..." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/demon_death.wav
say -v Alex "Forgive me... merciful... stars..." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/peasant_death.wav
say -v Alex "I have... failed... my quest..." -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/knight_death.wav

# Combat Grunts (Ah)
say -v Ralph "Ah!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/orc_hit.wav
say -v Whisper "Ah!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/demon_hit.wav
say -v Alex "Ah!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/peasant_hit.wav

# Knight (War Cry replacing sword swing)
say -v Alex "FOR GLORY!" -o tmp.aiff && ffmpeg -y -i tmp.aiff assets/audio/knight_attack.wav
# Knight hit
say -v Alex "Uh" -o tmp.aiff && ffmpeg -y -i tmp.aiff -af "volume=0.5" assets/audio/knight_hit.wav

rm tmp.aiff
