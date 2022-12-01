import api from '../utils/api';
import AsyncStorage from '@react-native-community/async-storage';

export default function logar(user) {
    return api.post('/session', user)
        .then(r => {
            console.alert(r.data);
            AsyncStorage.setItem("@token", r.data.token)
        })
        .catch(err => alert(err))
}