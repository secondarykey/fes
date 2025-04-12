import { useEffect, useState } from "react";

import cross from "./assets/crosshair.svg"
import back from "./assets/arrow-right-square-fill.svg"

const SVGButton = (props) => {

  const [svg,setSVG] = useState(<svg></svg>);

    useEffect( () => {
      var icon = svg;
      console.log("svgbutton")
      if ( props.name === "back" ) {
        icon = back;
      } else if ( props.name==="cross" ) {
        icon = cross;
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