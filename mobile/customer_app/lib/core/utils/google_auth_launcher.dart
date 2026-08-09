import 'google_auth_launcher_stub.dart'
    if (dart.library.html) 'google_auth_launcher_web.dart';

void openGoogleAuthWindow(String url) {
  launchGoogleOAuthWindow(url);
}
