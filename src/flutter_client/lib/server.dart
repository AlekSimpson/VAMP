import 'package:http/http.dart' as http;

Future<http.Response> getAudioList() async {
  return http.get(Uri.parse('http://localhost:8080/availableAudio'));
}