import { initializeApp } from 'firebase/app';
import { getAuth, GoogleAuthProvider, signInWithCredential, signInWithEmailAndPassword, createUserWithEmailAndPassword } from 'firebase/auth';
import { getFirestore, doc, getDoc, setDoc, onSnapshot } from 'firebase/firestore';

const firebaseConfig = {
    apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
    authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
    projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
    storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
    messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID,
    appId: import.meta.env.VITE_FIREBASE_APP_ID,
    measurementId: import.meta.env.VITE_FIREBASE_MEASUREMENT_ID
};

const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
export const db = getFirestore(app);
export const googleProvider = new GoogleAuthProvider();

// Used by system browser OAuth flow (LoginScreen)
export const signInWithGoogleIdToken = async (idToken) => {
    const credential = GoogleAuthProvider.credential(idToken);
    const result = await signInWithCredential(auth, credential);
    return result.user;
};

// Sign in with existing email and password
export const loginWithEmail = async (email, password) => {
    const result = await signInWithEmailAndPassword(auth, email.trim().toLowerCase(), password);
    return result.user;
};

// Register via Email/Password + Firestore
export const registerWithEmail = async (email, password) => {
    const result = await createUserWithEmailAndPassword(auth, email.trim().toLowerCase(), password);
    const user = result.user;
    const userRef = doc(db, 'users', user.uid);
    await setDoc(userRef, {
        email: email.trim().toLowerCase(),
        displayName: email.split('@')[0],
        is_enabled: false,
        created_at: new Date()
    }, { merge: true });
    return user;
};

export const getUserData = async (uid) => {
    const userRef = doc(db, 'users', uid);
    const userSnap = await getDoc(userRef);
    if (userSnap.exists()) {
        return userSnap.data();
    }
    return null;
};

export const createUserData = async (user) => {
    const userRef = doc(db, 'users', user.uid);
    await setDoc(userRef, {
        email: user.email,
        displayName: user.displayName,
        is_enabled: false,
        created_at: new Date()
    }, { merge: true });
};

export const onUserSnapshot = (uid, callback, onError) => {
    const userRef = doc(db, 'users', uid);
    return onSnapshot(userRef, (docSnap) => {
        if (docSnap.exists()) {
            callback(docSnap.data());
        } else {
            callback(null);
        }
    }, (error) => {
        console.error("Firestore onUserSnapshot error:", error);
        if (onError) onError(error);
    });
};

export const updateUserCookie = async (uid, platform, cookie) => {
    const userRef = doc(db, 'users', uid);
    const updateData = {};
    updateData[`cookies_${platform}`] = cookie;
    await setDoc(userRef, updateData, { merge: true });
};

export const updateUserConfig = async (uid, platform, config) => {
    const userRef = doc(db, 'users', uid);
    const updateData = {};
    updateData[`config_${platform}`] = config;
    await setDoc(userRef, updateData, { merge: true });
};
