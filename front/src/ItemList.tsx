import React from "react";

import {ToggleButton, ToggleButtonGroup, ListGroup} from "react-bootstrap";
import "bootstrap/dist/css/bootstrap.min.css";

import Item from "./Item";
import ItemProps from "./Item";
//import axios from "axios";
//import ClipLoader from 'react-spinners/ClipLoader';

type ItemListProps = {
    items : Array<ItemProps["props"]>;
    visibleCount : number;
}

function ItemList (props : ItemListProps) {
    // It's used to describe which tab is currently chosen (only visibleCount of items is display simultaenously).
    const [tab, setTab] = React.useState(0);

    function renderItems(items : Array<ItemProps["props"]>, visibleCount : number) {
	console.log(items);
	let offset = tab * visibleCount;
	return items.slice(offset, offset + visibleCount).map((item : ItemProps["props"], key) =>
			 <ListGroup.Item key={item.id} action><Item key={item.id} id={item.id} url={item.url} sleepTime={item.sleepTime} makeScreenshots={item.makeScreenshots} /></ListGroup.Item>)
    }

    function renderTabButtons(itemsCount : number, visibleCount : number) {
	let tabsCount = Math.ceil(itemsCount / visibleCount);
	let tabsIndexes = Array.from(Array(tabsCount).keys())
	return tabsIndexes.map((item) =>
			       <ToggleButton key={item} value={item}>{item}</ToggleButton>)
    }
    
  return (
      <div className="ItemList">
	<div className="Items m-2">
	  <ListGroup>
	    {renderItems(props.items, props.visibleCount)}
	  </ListGroup>
	</div>
	<div className="Controls m-2 mt-1">
	  <ToggleButtonGroup type="radio" onChange={(val : number) => setTab(val)} name="tab" defaultValue={0}>
	    {renderTabButtons(props.items.length, props.visibleCount)}
	  </ToggleButtonGroup>
	</div>
      </div>
  );
};

export default ItemList;
