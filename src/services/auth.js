// src/services/auth.js

import { GoogleAuthProvider, GithubAuthProvider, signInWithPopup, signInWithEmailAndPassword, signOut } from 'firebase/auth';
import { auth } from '../config/firebase';

// Google Login Function
export const googleLogin = async () => {
    const provider = new GoogleAuthProvider();
    try {
        const result = await signInWithPopup(auth, provider);
        // User info can be retrieved here
        return result.user;
    } catch (error) {
        console.error('Error during Google login:', error);
    }
};

// GitHub Login Function
export const githubLogin = async () => {
    const provider = new GithubAuthProvider();
    try {
        const result = await signInWithPopup(auth, provider);
        // User info can be retrieved here
        return result.user;
    } catch (error) {
        console.error('Error during GitHub login:', error);
    }
};

// Email/Password Login Function
export const emailPasswordLogin = async (email, password) => {
    try {
        const userCredential = await signInWithEmailAndPassword(auth, email, password);
        // User info can be retrieved here
        return userCredential.user;
    } catch (error) {
        console.error('Error during email/password login:', error);
    }
};

// Logout Function
export const logout = async () => {
    try {
        await signOut(auth);
        console.log('User signed out successfully');
    } catch (error) {
        console.error('Error signing out:', error);
    }
};

// Session Management (Check if user is logged in)
export const checkSession = () => {
    return new Promise((resolve) => {
        auth.onAuthStateChanged((user) => {
            resolve(user);
        });
    });
};
