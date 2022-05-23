import React, { Component } from "react";
import { Table, Label } from "semantic-ui-react";
import axios from "axios";

const endpoint = `http://localhost:8080`;

class LatestTxs extends Component {
  constructor(props) {
    super(props);
    this.state = {
      transactions: [],
    };
  }

  componentDidMount = () => {
    this.getTxs();
  };

  getTxs = async () => {
    const { blockNo } = this.props;
    if (blockNo) {
      console.log(blockNo)
      // get the block transaction
      let txsDetails = [];
      for (let j = 0; j < 10; j = j + 1) {
        const blockDetail = await axios.get(
            endpoint +
            `/block/${(blockNo-j)}`
        );

        const { transactions } = blockDetail.data.block.payload;
        // check if there is any transaction
        if (blockDetail.data.block.payload) {
          for (let i = 0; i < blockDetail.data.block.payload.length; i = i + 1) {
            const tx = blockDetail.data.block.payload[i];
            console.log(tx)
            txsDetails.push(
                <Table.Row key={i}>
                  <Table.Cell>
                    <Label color="blue">Tx</Label> {tx.signature}
                  </Table.Cell>
                  <Table.Cell>
                    From {tx.from} <br></br>
                    To {tx.to}
                  </Table.Cell>
                  <Table.Cell>
                    {" "}
                    <Label color="blue">AITU</Label> {tx.value}
                  </Table.Cell>
                </Table.Row>
            );
          }
        }

        this.setState({
          transactions: txsDetails,
        });
      }

    }

  };

  render() {
    return (
      <div>
        <Table fixed>
          <Table.Header>
            <Table.Row>
              <Table.Cell style={{ color: "#1d6fa5" }}>
                <h4> Latest Transactions</h4>
              </Table.Cell>
            </Table.Row>
          </Table.Header>
          <Table.Body>{this.state.transactions}</Table.Body>
        </Table>
      </div>
    );
  }
}

export default LatestTxs;