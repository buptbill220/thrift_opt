// Autogenerated by Thrift Compiler (0.9.2)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package echo

import (
	"bytes"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = bytes.Equal

var GoUnusedProtection__ int

type EchoReq struct {
	SeqId  int32              `thrift:"seq_id,1" json:"seq_id"`
	StrDat string             `thrift:"str_dat,2" json:"str_dat"`
	BinDat []byte             `thrift:"bin_dat,3" json:"bin_dat"`
	MDat   map[string]float64 `thrift:"m_dat,4" json:"m_dat"`
	T64    int64              `thrift:"t64,5" json:"t64"`
	Li32   []int32            `thrift:"li32,6" json:"li32"`
	T16    int16              `thrift:"t16,7" json:"t16"`
}

func NewEchoReq() *EchoReq {
	return &EchoReq{}
}

func (p *EchoReq) GetSeqId() int32 {
	return p.SeqId
}

func (p *EchoReq) GetStrDat() string {
	return p.StrDat
}

func (p *EchoReq) GetBinDat() []byte {
	return p.BinDat
}

func (p *EchoReq) GetMDat() map[string]float64 {
	return p.MDat
}

func (p *EchoReq) GetT64() int64 {
	return p.T64
}

func (p *EchoReq) GetLi32() []int32 {
	return p.Li32
}

func (p *EchoReq) GetT16() int16 {
	return p.T16
}
func (p *EchoReq) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return fmt.Errorf("%T read error: %s", p, err)
	}
	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return fmt.Errorf("%T field %d read error: %s", p, fieldId, err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		case 3:
			if err := p.ReadField3(iprot); err != nil {
				return err
			}
		case 4:
			if err := p.ReadField4(iprot); err != nil {
				return err
			}
		case 5:
			if err := p.ReadField5(iprot); err != nil {
				return err
			}
		case 6:
			if err := p.ReadField6(iprot); err != nil {
				return err
			}
		case 7:
			if err := p.ReadField7(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return fmt.Errorf("%T read struct end error: %s", p, err)
	}
	return nil
}

func (p *EchoReq) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return fmt.Errorf("error reading field 1: %s", err)
	} else {
		p.SeqId = v
	}
	return nil
}

func (p *EchoReq) ReadField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return fmt.Errorf("error reading field 2: %s", err)
	} else {
		p.StrDat = v
	}
	return nil
}

func (p *EchoReq) ReadField3(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadBinary(); err != nil {
		return fmt.Errorf("error reading field 3: %s", err)
	} else {
		p.BinDat = v
	}
	return nil
}

func (p *EchoReq) ReadField4(iprot thrift.TProtocol) error {
	_, _, size, err := iprot.ReadMapBegin()
	if err != nil {
		return fmt.Errorf("error reading map begin: %s", err)
	}
	tMap := make(map[string]float64, size)
	p.MDat = tMap
	for i := 0; i < size; i++ {
		var _key0 string
		if v, err := iprot.ReadString(); err != nil {
			return fmt.Errorf("error reading field 0: %s", err)
		} else {
			_key0 = v
		}
		var _val1 float64
		if v, err := iprot.ReadDouble(); err != nil {
			return fmt.Errorf("error reading field 0: %s", err)
		} else {
			_val1 = v
		}
		p.MDat[_key0] = _val1
	}
	if err := iprot.ReadMapEnd(); err != nil {
		return fmt.Errorf("error reading map end: %s", err)
	}
	return nil
}

func (p *EchoReq) ReadField5(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return fmt.Errorf("error reading field 5: %s", err)
	} else {
		p.T64 = v
	}
	return nil
}

func (p *EchoReq) ReadField6(iprot thrift.TProtocol) error {
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return fmt.Errorf("error reading list begin: %s", err)
	}
	tSlice := make([]int32, 0, size)
	p.Li32 = tSlice
	for i := 0; i < size; i++ {
		var _elem2 int32
		if v, err := iprot.ReadI32(); err != nil {
			return fmt.Errorf("error reading field 0: %s", err)
		} else {
			_elem2 = v
		}
		p.Li32 = append(p.Li32, _elem2)
	}
	if err := iprot.ReadListEnd(); err != nil {
		return fmt.Errorf("error reading list end: %s", err)
	}
	return nil
}

func (p *EchoReq) ReadField7(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI16(); err != nil {
		return fmt.Errorf("error reading field 7: %s", err)
	} else {
		p.T16 = v
	}
	return nil
}

func (p *EchoReq) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("EchoReq"); err != nil {
		return fmt.Errorf("%T write struct begin error: %s", p, err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := p.writeField3(oprot); err != nil {
		return err
	}
	if err := p.writeField4(oprot); err != nil {
		return err
	}
	if err := p.writeField5(oprot); err != nil {
		return err
	}
	if err := p.writeField6(oprot); err != nil {
		return err
	}
	if err := p.writeField7(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return fmt.Errorf("write field stop error: %s", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return fmt.Errorf("write struct stop error: %s", err)
	}
	return nil
}

func (p *EchoReq) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("seq_id", thrift.I32, 1); err != nil {
		return fmt.Errorf("%T write field begin error 1:seq_id: %s", p, err)
	}
	if err := oprot.WriteI32(int32(p.SeqId)); err != nil {
		return fmt.Errorf("%T.seq_id (1) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 1:seq_id: %s", p, err)
	}
	return err
}

func (p *EchoReq) writeField2(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("str_dat", thrift.STRING, 2); err != nil {
		return fmt.Errorf("%T write field begin error 2:str_dat: %s", p, err)
	}
	if err := oprot.WriteString(string(p.StrDat)); err != nil {
		return fmt.Errorf("%T.str_dat (2) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 2:str_dat: %s", p, err)
	}
	return err
}

func (p *EchoReq) writeField3(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("bin_dat", thrift.STRING, 3); err != nil {
		return fmt.Errorf("%T write field begin error 3:bin_dat: %s", p, err)
	}
	if err := oprot.WriteBinary(p.BinDat); err != nil {
		return fmt.Errorf("%T.bin_dat (3) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 3:bin_dat: %s", p, err)
	}
	return err
}

func (p *EchoReq) writeField4(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("m_dat", thrift.MAP, 4); err != nil {
		return fmt.Errorf("%T write field begin error 4:m_dat: %s", p, err)
	}
	if err := oprot.WriteMapBegin(thrift.STRING, thrift.DOUBLE, len(p.MDat)); err != nil {
		return fmt.Errorf("error writing map begin: %s", err)
	}
	for k, v := range p.MDat {
		if err := oprot.WriteString(string(k)); err != nil {
			return fmt.Errorf("%T. (0) field write error: %s", p, err)
		}
		if err := oprot.WriteDouble(float64(v)); err != nil {
			return fmt.Errorf("%T. (0) field write error: %s", p, err)
		}
	}
	if err := oprot.WriteMapEnd(); err != nil {
		return fmt.Errorf("error writing map end: %s", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 4:m_dat: %s", p, err)
	}
	return err
}

func (p *EchoReq) writeField5(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("t64", thrift.I64, 5); err != nil {
		return fmt.Errorf("%T write field begin error 5:t64: %s", p, err)
	}
	if err := oprot.WriteI64(int64(p.T64)); err != nil {
		return fmt.Errorf("%T.t64 (5) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 5:t64: %s", p, err)
	}
	return err
}

func (p *EchoReq) writeField6(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("li32", thrift.LIST, 6); err != nil {
		return fmt.Errorf("%T write field begin error 6:li32: %s", p, err)
	}
	if err := oprot.WriteListBegin(thrift.I32, len(p.Li32)); err != nil {
		return fmt.Errorf("error writing list begin: %s", err)
	}
	for _, v := range p.Li32 {
		if err := oprot.WriteI32(int32(v)); err != nil {
			return fmt.Errorf("%T. (0) field write error: %s", p, err)
		}
	}
	if err := oprot.WriteListEnd(); err != nil {
		return fmt.Errorf("error writing list end: %s", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 6:li32: %s", p, err)
	}
	return err
}

func (p *EchoReq) writeField7(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("t16", thrift.I16, 7); err != nil {
		return fmt.Errorf("%T write field begin error 7:t16: %s", p, err)
	}
	if err := oprot.WriteI16(int16(p.T16)); err != nil {
		return fmt.Errorf("%T.t16 (7) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 7:t16: %s", p, err)
	}
	return err
}

func (p *EchoReq) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("EchoReq(%+v)", *p)
}

type EchoRsp struct {
	Status int32  `thrift:"status,1" json:"status"`
	Msg    string `thrift:"msg,2" json:"msg"`
}

func NewEchoRsp() *EchoRsp {
	return &EchoRsp{}
}

func (p *EchoRsp) GetStatus() int32 {
	return p.Status
}

func (p *EchoRsp) GetMsg() string {
	return p.Msg
}
func (p *EchoRsp) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return fmt.Errorf("%T read error: %s", p, err)
	}
	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return fmt.Errorf("%T field %d read error: %s", p, fieldId, err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return fmt.Errorf("%T read struct end error: %s", p, err)
	}
	return nil
}

func (p *EchoRsp) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return fmt.Errorf("error reading field 1: %s", err)
	} else {
		p.Status = v
	}
	return nil
}

func (p *EchoRsp) ReadField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return fmt.Errorf("error reading field 2: %s", err)
	} else {
		p.Msg = v
	}
	return nil
}

func (p *EchoRsp) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("EchoRsp"); err != nil {
		return fmt.Errorf("%T write struct begin error: %s", p, err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return fmt.Errorf("write field stop error: %s", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return fmt.Errorf("write struct stop error: %s", err)
	}
	return nil
}

func (p *EchoRsp) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("status", thrift.I32, 1); err != nil {
		return fmt.Errorf("%T write field begin error 1:status: %s", p, err)
	}
	if err := oprot.WriteI32(int32(p.Status)); err != nil {
		return fmt.Errorf("%T.status (1) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 1:status: %s", p, err)
	}
	return err
}

func (p *EchoRsp) writeField2(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("msg", thrift.STRING, 2); err != nil {
		return fmt.Errorf("%T write field begin error 2:msg: %s", p, err)
	}
	if err := oprot.WriteString(string(p.Msg)); err != nil {
		return fmt.Errorf("%T.msg (2) field write error: %s", p, err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 2:msg: %s", p, err)
	}
	return err
}

func (p *EchoRsp) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("EchoRsp(%+v)", *p)
}
