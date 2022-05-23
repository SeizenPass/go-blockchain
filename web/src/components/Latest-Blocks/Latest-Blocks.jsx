import React, { Component } from "react";
import { Table, Label } from "semantic-ui-react";

import axios from "axios";

const endpoint = `http://localhost:8080`;
function timeConverter(UNIX_timestamp){
    var a = new Date(UNIX_timestamp * 1000);
    var months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
    var year = a.getFullYear();
    var month = months[a.getMonth()];
    var date = a.getDate();
    var hour = a.getHours();
    var min = a.getMinutes();
    var sec = a.getSeconds();
    var time = date + ' ' + month + ' ' + year + ' ' + hour + ':' + min + ':' + sec ;
    return time;
}
class LatestBlocks extends Component {
  constructor(props) {
    super(props);
    this.state = {
      blocks: [],
    };
  }

  componentDidMount = () => {
    this.getBlocks();
  };

  getBlocks = async () => {
    const { latestBlock } = this.props;

    let blocks = [];

    // check if latest blocks
    if (latestBlock) {
      for (let i = 0; i < 10    ; i = i + 1) {
        // get the block transaction
        const blockDetail = await axios.get(
          endpoint +
            `/block/${(
              latestBlock - i
            )}`
        );

        let { result } = blockDetail.data;
        blocks.push(
          <Table.Row key={i}>
            <Table.Cell>
              <Label color="blue">Bk</Label> {latestBlock - i}
            </Table.Cell>
            <Table.Cell>
              Miner {blockDetail.data.block.header.miner} <br></br>
              Txs {blockDetail.data.block.payload.length}
            </Table.Cell>
            <Table.Cell>
              <Label color="blue">Time </Label> {timeConverter(blockDetail.data.block.header.time)}
            </Table.Cell>
          </Table.Row>
        );

        this.setState({
          blocks: blocks,
        });
      }
    }
  };

  render() {
    return (
      <Table fixed>
        <Table.Header>
          <Table.Row>
            <Table.Cell style={{ color: "#1d6fa5" }}>
              <h4>Latest Blocks</h4>
            </Table.Cell>
          </Table.Row>
        </Table.Header>

        <Table.Body>{this.state.blocks}</Table.Body>
      </Table>
    );
  }
}

export default LatestBlocks;