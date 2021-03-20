// This example demos the TextField/SelectableText widget and keyboard
// integration with the go-flutter text backend

import 'package:flutter/material.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter Code Sample for testing text input',
      theme: ThemeData(
        // If the host is missing some fonts, it can cause the
        // text to not be rendered or worse the app might crash.
        fontFamily: 'Roboto',
        primarySwatch: Colors.blue,
      ),
      home: MyStatefulWidget(),
    );
  }
}

class MyStatefulWidget extends StatefulWidget {
  MyStatefulWidget({Key key}) : super(key: key);

  @override
  _MyStatefulWidgetState createState() => _MyStatefulWidgetState();
}

class _MyStatefulWidgetState extends State<MyStatefulWidget> {
  FocusNode myFocus = FocusNode();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('O jogo'),
        backgroundColor: Color.fromRGBO(0, 0, 0, 1),
      ),
      body: Center(child:
              Column(children: [
                Text("O Jogo", style: TextStyle(fontSize: 69)),
                Text("Você perdeu")
              ], mainAxisAlignment: MainAxisAlignment.center,
      )));
  }
}
