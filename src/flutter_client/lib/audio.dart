import 'package:sprintf/sprintf.dart';

class Audio {
  final String name;
  final String author;
  final int    durationInSeconds;

  const Audio({
    required this.name,
    required this.author,
    required this.durationInSeconds,
  });

  @override
  String toString() {
    return sprintf("Audio [name: %s, author: %s, duration (sec): %d]", [name, author, durationInSeconds]);
  }

  String stripMp3Tag() {
    return author.split(".")[0];
  }

  String encodeAudioName() {
    String retval = sprintf("%s_%s", [name, stripMp3Tag()]);
    String upper = retval.toUpperCase();
    return upper.replaceAll(" ", "-");
  }

  factory Audio.fromJSON(Map<String, dynamic> json) {
    return switch (json) {
      {
        'name': String name,
        'author': String author,
        'duration': int duration,
      } =>
        Audio(author: author, name: name, durationInSeconds: duration),
      _ => throw const FormatException('Failed to load album.'),
    };
  }

  factory Audio.empty() {
    return const Audio(author: "", name: "", durationInSeconds: 0);
  }
}