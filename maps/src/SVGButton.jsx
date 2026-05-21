import { useEffect, useState } from "react";

import cross from "./assets/crosshair.svg"
import back from "./assets/arrow-right-square-fill.svg"
import update from "./assets/arrow-clockwise.svg"
import terminal from "./assets/terminal.svg"

const SVGButton = (props) => {

  const [svg,setSVG] = useState(<svg></svg>);

    useEffect( () => {
      var icon = svg;
      if ( props.name === "back" ) {
        icon = back;
      } else if ( props.name==="cross" ) {
        icon = cross;
      } else if ( props.name==="update" ) {
        icon = update;
      } else if ( props.name==="terminal" ) {
        icon = terminal;
      }
      setSVG(icon);
    },[])

    return (<>
      <button {...props} className={"svgBtn " + (props.className === undefined ? "" : props.className)}>
        <object type="image/svg+xml" data={svg} width={props.size} height={props.size}/>
      </button>
    </>)
}
export default SVGButton;