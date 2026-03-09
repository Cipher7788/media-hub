import React from 'react';
import './Dashboard.css';

const Dashboard = () => {
    return (
        <div className="dashboard">
            <h1>User Dashboard</h1>
            <section className="statistics">
                <h2>Statistics</h2>
                <p>Here will be displayed user statistics.</p>
            </section>
            <section className="recent-activity">
                <h2>Recent Activity</h2>
                <p>Latest activities will be listed here.</p>
            </section>
            <section className="collections-overview">
                <h2>Collections Overview</h2>
                <p>Overview of user collections.</p>
            </section>
            <section className="favorite-items">
                <h2>Favorite Items</h2>
                <p>A list of the user's favorite items.</p>
            </section>
            <section className="watchlist-progress">
                <h2>Watchlist Progress</h2>
                <p>Progress on the user's watchlist items.</p>
            </section>
            <section className="personal-ratings">
                <h2>Personal Ratings</h2>
                <p>User's personal ratings can be displayed here.</p>
            </section>
        </div>
    );
};

export default Dashboard;
