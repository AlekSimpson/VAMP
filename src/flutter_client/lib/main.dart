// ignore_for_file: avoid_print

import 'dart:convert';

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:audioplayers/audioplayers.dart';
import 'package:flutter/services.dart';
import 'package:sprintf/sprintf.dart';

import 'server.dart';
import 'audio.dart';

void main() {
  runApp(const MyApp());
}

void printf(String message, List<dynamic> args) {
  print(sprintf(message, args));
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'VAMP',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color.fromARGB(255, 65, 122, 214)),
        useMaterial3: true,
      ),
      home: const MyHomePage(title: 'VAMP'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  // TOPOFCLASS
  late AudioPlayer player;
  bool isPlaying = false;
  bool isLooping = false;
  bool isShuffle = false;
  String currentlyPlaying = "";
  double currentVolume = 0.01;
  double playerPosition = 0.0;
  late List<Audio> cachedAudio;
  late List<String> cachedAudioNames; // temporary cache will be overwritten and changed while program plays

  @override
  void dispose() {
    player.dispose();
    super.dispose();
  }

  Future<List<Audio>> loadAudio() async {
    final response = await getAudioList();
    List<Audio> retval = [];
    if (response.statusCode == 200) {
      // decode the reponse into Map<String, dynamic> object
      final decodedResponse = jsonDecode(response.body);
      // convert map into list of Audio objects
      for (int i = 0; i < decodedResponse.length; i++) {
        retval.add(Audio.fromJSON(decodedResponse[i]));
      }

      return retval;
    }
    else {
      // give http error code
      throw Exception("failed with error http code: ${response.statusCode}");
    }
  }

  @override
  void initState() {
    super.initState();
    player = AudioPlayer();
    player.setReleaseMode(ReleaseMode.stop);

    player.onPlayerComplete.listen((event) {
      if (isLooping) {
        handlePlaySongEvent(currentlyPlaying);
      }
      else if (isShuffle) {
        String nextSong = cachedAudioNames[0];
        setState(() { 
          cachedAudioNames.remove(nextSong); 
          currentlyPlaying = nextSong;
        });
        handlePlaySongEvent(nextSong);
      }
      else {
        setState(() {
         isPlaying = false;         
        });
      }
    });
  }

  Future<void> handlePlaySongEvent(String name) async {
    if (currentlyPlaying == name) {
      await player.release();
    }

    currentlyPlaying = name;
    String url = sprintf("http://localhost:8080/audio?file=%s", [name]);

    setState(() {isPlaying = true;});
    await player.stop();
    await player.setSourceUrl(url);
    player.setVolume(currentVolume);
    await player.resume();

    //player.onPositionChanged.listen((position) {
    //  setState(() {
    //    player.getDuration().then((duration) {
    //      if (duration == null) {
    //        playerPosition = position.inSeconds / duration!.inSeconds;
    //      }
    //    });
    //  });
    //});
  }

  void toggleLoop() async {
    setState(() {
      isLooping = !isLooping;
    });
  }

  List<String> extractNames() {
    List<String> retval = [];
    for (Audio a in cachedAudio) {
      retval.add(a.encodeAudioName());
    }
    return retval;
  }

  void toggleShuffle() {
    setState(() {
      isShuffle = !isShuffle;

      if (player.source == null) {
        List<String> audioNames = extractNames();
        audioNames.remove(currentlyPlaying);
        audioNames.shuffle();
        cachedAudioNames = audioNames;
      }
    });
  }

  void toggleAudioPlayback() {
    if (player.source == null) {
      return;
    }

    setState(() {
      if (isPlaying) {
        player.pause();
        isPlaying = false;
      }
      else {
        player.resume();
        isPlaying = true;
      }
    });
  }

  // MARK: Widget Builders
  FloatingActionButton buildItem(Audio audio) {
    List<String> delimited = audio.author.split('.'); 
    String stripMp2Tag = delimited[0];

    return FloatingActionButton(
      onPressed: () {
        handlePlaySongEvent(audio.encodeAudioName());
      },
      tooltip: 'play song',
      elevation: 1.5,
      child: ListTile(title: Text(audio.name), subtitle: Text(stripMp2Tag)),
    );
  }

  Center buildCenterBodyContent() {
    return Center(
        child: FutureBuilder<List<Audio>>(
          future: loadAudio(),
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting && !snapshot.hasData) {
              return const Center(child: CircularProgressIndicator());
            }
            else if (snapshot.hasError) {
              return const Center(child: Text("Error loading data"));
            }
            else if (!snapshot.hasData || snapshot.data!.isEmpty) {
              return const Center(child: Text("No available data"));
            }
            else {
              List<Audio> audios = snapshot.data!;
              cachedAudio = audios;
              return ListView.separated(
                itemCount: audios.length,
                itemBuilder: (context, index) {
                  return buildItem(audios[index]);
                },
                separatorBuilder: (context, index) {
                  return const SizedBox(height: 5.0);
                }
              );
            }
          },
        ),
      );
  }

  Container buildLeftBar(ColorScheme scheme) {
    return Container(
      width: 175, // Adjust the width as needed
      color: scheme.inversePrimary,
      child: Column(children: [
          ListTile(
            leading: const Icon(Icons.audiotrack),
            title: const Text('Default'),
            onTap: () {},
          ),
          ListTile(
            leading: const Icon(Icons.audiotrack),
            title: const Text('Playlist 1'),
            onTap: () {},
          ),
          ListTile(
            leading: const Icon(Icons.audiotrack),
            title: const Text('Playlist 2'),
            onTap: () {},
          ),
        ],
      ),
    );
  }

  // WIDGETBUILDER
  @override
  Widget build(BuildContext context) {
    var scheme = Theme.of(context).colorScheme;
    var style = TextButton.styleFrom(foregroundColor: scheme.primary);

    return Scaffold(
      appBar: AppBar(
        backgroundColor: scheme.inversePrimary,
        actions: [
          IconButton(icon: const Icon(Icons.all_inclusive), color: isLooping ? scheme.onPrimary : scheme.primary, onPressed: toggleLoop),
          IconButton(icon: isPlaying ? const Icon(Icons.pause) : const Icon(Icons.play_arrow), color: scheme.primary, onPressed: toggleAudioPlayback),
          IconButton(icon: const Icon(Icons.shuffle), color: isShuffle ? scheme.onPrimary : scheme.primary, onPressed: toggleShuffle),

          // quit button
          TextButton(style: style, onPressed: () {
            SystemChannels.platform.invokeMethod('SystemNavigator.pop');
          }, child: const Text("Quit")),
        ],
        title: Text(widget.title),
      ),
      body: Row(children: [buildLeftBar(scheme), Expanded(child: buildCenterBodyContent())],),
    );
  }
}
