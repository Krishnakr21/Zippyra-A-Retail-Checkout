// ignore: avoid_web_libraries_in_flutter
import 'dart:html' as html;

void launchGoogleOAuthWindow(String url) {
  html.window.open(url, 'GoogleSignInWindow', 'width=520,height=640,top=100,left=100,scrollbars=yes,status=yes');
}
