import React, {useState, useEffect} from 'react';
import {View, TextInput, Text, StyleSheet, TouchableOpacity} from 'react-native';
import logar from '../services/userService'

import {
  widthPercentageToDP as wp,
  heightPercentageToDP as hp
} from "../utils/responsive";
import { navigate } from '../utils/navigate';
import AsyncStorage from '@react-native-community/async-storage';


export default function Login() {

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")

  useEffect(()=>{
    AsyncStorage
      .getItem("@access_token")
      .then(token => navigate("PedidoPage"))
  },[])

  const validate = () => {
    
    if(!email || !email.trim){
      alert("Email inválido")
      return false
    }
    
    if(!password || !password.trim){
      alert("Senha inválida")
      return false
    }

    return true
  }

  const handleLogin = () => {
    if(validate())
      logar({email:email, password:password})
    
}

  return (
    <View style={styles.page}>
      <View style={styles.title}>
        <Text style={styles.textTitle}>Lanchonete</Text>
        <Text style={styles.textTitle}>do</Text>
        <Text style={styles.textTitleZ}>ZAMIGOS</Text>
      </View>
      
      <View style={styles.body}>

        <View style={styles.containerInputs}>
          <Text>Usuário</Text>
          <TextInput style={styles.inputs} onChangeText = {setEmail}/>
        </View>
        
        <View style={styles.containerInputs}>
          <Text>Senha</Text>
          <TextInput style={styles.inputs} onChangeText = {setPassword}/>
        </View>

        <View>
          <TouchableOpacity style={styles.entrar} onPress={handleLogin}>
            <Text>Entrar</Text>
          </TouchableOpacity>
        </View>

      </View>
    </View>
  );
}


const styles = StyleSheet.create({
  page:{
    flex:1,
    backgroundColor:'yellow',
  },
  title:{
    alignItems:'center',
    justifyContent:"center",
    flex:2
  },
  textTitle:{
    fontSize:45
  },
  body:{
    flex:5,
    justifyContent:"center",
    alignItems:'center'
  },
  containerInputs:{
    justifyContent:"center",
    marginVertical:12
  },
  inputs:{
    backgroundColor:'white',
    width:wp("90%"),
    borderRadius:10
  },
  entrar:{
    borderColor:"grey",
    borderWidth:.5,
    borderRadius:9,
    paddingHorizontal:25,
    paddingVertical:10,
    backgroundColor:"#d1fffe",
    width:wp("90%"),
    alignItems:'center'
  },
  textTitleZ:{
    fontSize:45,
    textShadowColor:'green',
    textShadowRadius:5
},

})