import React from 'react';
import PropTypes from 'prop-types';

const MediaDetail = ({ title, description, ratings, trailers, cast, reviews, onAddToWatchlist, onAddToFavorites }) => {
    return (
        <div className="media-detail">
            <h1>{title}</h1>
            <p>{description}</p>
            <p><strong>Ratings:</strong> {ratings}</p>
            <h2>Trailers</h2>
            <div className="trailers">
                {trailers.map((trailer, index) => (
                    <iframe key={index} src={trailer} title={`Trailer ${index + 1}`} frameBorder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen></iframe>
                ))}
            </div>
            <h2>Cast</h2>
            <ul>
                {cast.map((member, index) => (
                    <li key={index}>{member}</li>
                ))}
            </ul>
            <h2>Reviews</h2>
            <ul>
                {reviews.map((review, index) => (
                    <li key={index}>{review}</li>
                ))}
            </ul>
            <button onClick={onAddToWatchlist}>Add to Watchlist</button>
            <button onClick={onAddToFavorites}>Add to Favorites</button>
        </div>
    );
};

MediaDetail.propTypes = {
    title: PropTypes.string.isRequired,
    description: PropTypes.string.isRequired,
    ratings: PropTypes.string,
    trailers: PropTypes.arrayOf(PropTypes.string).isRequired,
    cast: PropTypes.arrayOf(PropTypes.string).isRequired,
    reviews: PropTypes.arrayOf(PropTypes.string).isRequired,
    onAddToWatchlist: PropTypes.func.isRequired,
    onAddToFavorites: PropTypes.func.isRequired,
};

export default MediaDetail;