import React from 'react';
import './ExploreMenu.css';
import { menu_list } from '../../assets/assets';

const ExploreMenu = ({ category, setCategory }) => {
    return (
        <div className='explore-menu' id='explore-menu'>
            <h1>Explore Our Menu</h1>
            <p className='explore-menu-text'>Indulge in a variety of mouthwatering pizzas, crafted with the freshest ingredients and bold flavors. From timeless classics to unique creations, there's something for everyone. Treat yourself to a dining experience that's as satisfying as it is delicious </p>
            <div className="explore-menu-list">
                {menu_list.map((item, index) => (
                    <div key={index} className="explore-menu-list-items" onClick={() => setCategory(prev => prev === item.menu_name ? "All" : item.menu_name)}>
                        <img src={item.menu_image} alt="" className={`menu-image ${category === item.menu_name ? 'active' : ''}`} />
                        <p>{item.menu_name}</p>
                    </div>
                ))}
            </div>
            <hr />
        </div>
    );
};

export default ExploreMenu;
