import { initializeApp } from 'firebase/app';
import { getAuth, GoogleAuthProvider, signInWithPopup } from 'firebase/auth';
import { getFirestore, doc, getDoc, setDoc, onSnapshot } from 'firebase/firestore';

const firebaseConfig = {
    apiKey: "AIzaSyA0xDczfRxVE7QbVKQpm2EkM8pleF52KOk",
    authDomain: "jdsa-ictg.firebaseapp.com",
    projectId: "jdsa-ictg",
    storageBucket: "jdsa-ictg.firebasestorage.app",
    messagingSenderId: "775589919465",
    appId: "1:775589919465:web:64fca7c0deab9f001dce78",
    measurementId: "G-2E3XBLX0CJ"
};

const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
export const db = getFirestore(app);
export const googleProvider = new GoogleAuthProvider();

export const loginWithGoogle = async () => {
    try {
        const result = await signInWithPopup(auth, googleProvider);
        return result.user;
    } catch (error) {
        console.error("Error signing in with Google:", error);
        throw error;
    }
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

export const onUserSnapshot = (uid, callback) => {
    const userRef = doc(db, 'users', uid);
    return onSnapshot(userRef, (docSnap) => {
        if (docSnap.exists()) {
            callback(docSnap.data());
        }
    });
};

export const updateUserCookie = async (uid, platform, cookie) => {
    const userRef = doc(db, 'users', uid);
    const updateData = {};
    updateData[`cookies_${platform}`] = cookie;
    await setDoc(userRef, updateData, { merge: true });
};
