import React, { Component } from "react";
import axios from "axios";
import "./eth-overview.css";
import { Card, Grid, Icon } from "semantic-ui-react";
import LatestBlocks from "../Latest-Blocks/index";
import LatestTxs from "../Latest-Txs/index";

const endpoint = `http://localhost:8080`;

class EthOverview extends Component {
  constructor() {
    super();
    this.state = {
      block_hash: "",
      block_number: 0,
    };
  }

  async componentDidMount() {
    // get the latest block number
    const latestBlock = await axios.get(
      endpoint + `/node/status`
    );
    this.setState({
      block_hash: parseInt(latestBlock.data.block_hash),
      block_number: latestBlock.data.block_number, // save block no in hex
    });
  }

  getLatestBlocks = () => {
    if (this.state.block_number) {
      return <LatestBlocks latestBlock={this.state.block_number}></LatestBlocks>;
    }
  };

  getLatestTxs = () => {
    if (this.state.block_number) {
      return <LatestTxs blockNo={this.state.block_number}></LatestTxs>;
    }
  };

  render() {
    return (
      <div>
        <Grid>
          <Grid.Row>
            <Grid.Column width={4}>
              <Card>
                <Card.Content>
                  <Card.Header style={{ color: "#1d6fa5" }}>
                    <Icon name="list alternate outline"></Icon> LATEST BLOCK
                  </Card.Header>
                  <Card.Description textAlign="left">
                    <Icon name="square"></Icon> {this.state.block_hash}
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
          </Grid.Row>
        </Grid>

        <Grid divided="vertically">
          <Grid.Row columns={2}>
            <Grid.Column>{this.getLatestBlocks()}</Grid.Column>
            <Grid.Column>{this.getLatestTxs()}</Grid.Column>
          </Grid.Row>
        </Grid>
      </div>
    );
  }
}

export default EthOverview;