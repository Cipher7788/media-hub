import {
  createUserWithEmailAndPassword,
  GithubAuthProvider,
  GoogleAuthProvider,
  signInWithEmailAndPassword,
  signInWithPopup,
  signOut,
} from 'firebase/auth';
import { auth } from '../config/firebase';

export async function googleLogin() {
  const { user } = await signInWithPopup(auth, new GoogleAuthProvider());
  return user;
}

export async function githubLogin() {
  const { user } = await signInWithPopup(auth, new GithubAuthProvider());
  return user;
}

export async function emailPasswordLogin(email, password) {
  const { user } = await signInWithEmailAndPassword(auth, email, password);
  return user;
}

export async function register(email, password) {
  const { user } = await createUserWithEmailAndPassword(auth, email, password);
  return user;
}

export function logout() {
  return signOut(auth);
}

export async function checkSession() {
  await auth.authStateReady();
  return auth.currentUser;
}
