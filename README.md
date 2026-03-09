# MediaHub 🎬📺📖💥

A personal media collection web app for comics, anime, movies, and books. Search, organize, and rate your favorite media with real-time sync across devices.

## ✨ Features

- **Search & Discover**: Search movies, anime, comics, and books using public APIs
- **User Authentication**: Secure login with Firebase Authentication
- **Personal Collections**: 
  - Save favorites
  - Create watchlists
  - Rate and review media
- **Responsive Gallery**: Beautiful grid layout with images and descriptions
- **Real-time Database**: Firestore sync for seamless cross-device experience
- **Modern UI**: Built with React for smooth interactions

## 🛠️ Tech Stack

- **Frontend**: React.js with React Router
- **Backend**: Firebase (Authentication, Firestore)
- **APIs**: 
  - TMDB (Movies & TV Shows)
  - Jikan (Anime)
  - Comic Vine (Comics & Manga)
- **Styling**: CSS3 with responsive design
- **Deployment**: Firebase Hosting / GitHub Pages

## 🚀 Quick Start

### Prerequisites
- Node.js 14+ and npm
- Firebase account
- API keys for:
  - [TMDB](https://www.themoviedb.org/settings/api)
  - [Comic Vine](https://comicvine.gamespot.com/api/)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/Cipher7788/media-hub.git
cd media-hub
```

2. **Install dependencies**
```bash
npm install
```

3. **Set up environment variables**
```bash
cp .env.example .env.local
```

Edit `.env.local` with your Firebase and API credentials:
```
REACT_APP_FIREBASE_API_KEY=your_key
REACT_APP_FIREBASE_AUTH_DOMAIN=your_domain
REACT_APP_FIREBASE_PROJECT_ID=your_project_id
REACT_APP_FIREBASE_STORAGE_BUCKET=your_bucket
REACT_APP_FIREBASE_MESSAGING_SENDER_ID=your_sender_id
REACT_APP_FIREBASE_APP_ID=your_app_id

REACT_APP_TMDB_API_KEY=your_tmdb_key
```

4. **Start development server**
```bash
npm start
```

The app will open at `http://localhost:3000`

## 📁 Project Structure

```
media-hub/
├── public/
├── src/
│   ├── components/
│   │   ├── Navigation.jsx
│   │   └── Navigation.css
│   ├── config/
│   │   └── firebase.js
│   ├── hooks/
│   │   └── useAuth.js
│   ├── pages/
│   │   ├── Home.jsx
│   │   ├── Home.css
│   │   ├── Search.jsx
│   │   └── Search.css
│   ├── services/
│   │   ├── api.js
│   │   ├── auth.js
│   │   └── firestore.js
│   ├── App.jsx
│   ├── App.css
│   └── index.js
├── .env.example
├── package.json
└── README.md
```

## 🎯 Usage

### Search for Media
1. Click on a category or use the search bar
2. Enter a query (movie title, anime name, etc.)
3. Browse results and click to view details

### Save to Collections
- **Add to Favorites**: Click the heart icon
- **Add to Watchlist**: Click the clock icon
- **Rate & Review**: Give a rating and write a review

### View Your Dashboard
- Go to your dashboard to see all your collections
- Track watching progress
- Compare ratings with other users

## 🔐 Firebase Setup

1. Create a Firebase project at [firebase.google.com](https://firebase.google.com)
2. Enable:
   - Authentication (Email/Password)
   - Firestore Database
3. Copy credentials to `.env.local`

### Firestore Rules

Add these security rules in Firebase Console:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /users/{userId} {
      allow read, write: if request.auth.uid == userId;
      match /{document=**} {
        allow read, write: if request.auth.uid == userId;
      }
    }
  }
}
```

## 📦 Building for Production

```bash
npm run build
```

## 🚀 Deployment

### Firebase Hosting
```bash
npm install -g firebase-tools
firebase login
firebase deploy
```

### GitHub Pages
```bash
npm run deploy
```

## 🎨 Customization

- Modify colors in CSS files (default: purple gradient)
- Add more APIs by extending `src/services/api.js`
- Create new pages in `src/pages/`
- Add more components in `src/components/`

## 🤖 Future Enhancements

- [ ] Recommendations engine
- [ ] Social features (following, sharing)
- [ ] Reviews and ratings community
- [ ] Mobile app (React Native)
- [ ] AI-powered tagging and suggestions
- [ ] Advanced filtering and sorting
- [ ] Export collection as PDF
- [ ] Dark mode

## 📝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see LICENSE file for details.

## 🤝 Support

Found a bug? Have a suggestion? 
- [Open an issue](https://github.com/Cipher7788/media-hub/issues)
- [Start a discussion](https://github.com/Cipher7788/media-hub/discussions)

## 🙏 Acknowledgments

- TMDB for movie database
- Jikan for anime data
- Comic Vine for comics
- Firebase for backend infrastructure
- React community for amazing tools

---

Made with ❤️ by [Cipher7788](https://github.com/Cipher7788)

**[Live Demo](#)** • **[Report Bug](https://github.com/Cipher7788/media-hub/issues)** • **[Request Feature](https://github.com/Cipher7788/media-hub/discussions)**